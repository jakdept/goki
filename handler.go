package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ajg/form"
	"github.com/blevesearch/bleve"
)

// Fields is a standard handler that pulls the first folder of the response, and
//  lists that topic or author. If there is none, it falls back to listing all
//  topics or authors with the fallback template.
type Fields struct {
	c ServerSection
	i *Index
}

func (h Fields) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fields := strings.SplitN(r.URL.Path, "/", 3)

	if len(fields) < 3 || fields[1] == "" {
		// to do if a field was not given
		h.i.FallbackSearchResponse(w, h.c.FallbackTemplate)
		return
	} else {
		// to be done if a field was given
		response, err := h.i.ListFieldValues(h.c.Default, fields[0], 100, 1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		modifiedSearchResponse := struct {
			SearchResponse
			AllFields []string
		}{
			SearchResponse: response,
			AllFields:      []string{},
		}

		err = allTemplates.ExecuteTemplate(w, h.c.Template, modifiedSearchResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// FuzzySearch is a normal search format - it should provide a point and click
//  interface to allow searching.
type FuzzySearch struct {
	c ServerSection
	i *Index
}

//FuzzySearchValues gives a standard structure to decode and pass to FuzzySearch
type FuzzySearchValues struct {
	s        string   `form:"s"`
	topics   []string `form:"topic"`
	authors  []string `form:"author"`
	page     int      `form:"page"`
	pageSize int      `form:"pageSize"`
}

func (h FuzzySearch) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// to do if a field was not given
		h.i.FallbackSearchResponse(w, h.c.FallbackTemplate)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var values FuzzySearchValues
	err = form.DecodeValues(&values, r.Form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := h.i.FuzzySearch(values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = allTemplates.ExecuteTemplate(w, h.c.Template, results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// QuerySearch is a handler that uses a custom search format to do custom queries.
type QuerySearch struct {
	c ServerSection
	i *Index
}

func (h QuerySearch) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values := struct {
		s        string `form:"s"`
		page     int    `form:"page"`
		pageSize int    `form:"pageSize"`
	}{}

	if r.Method != http.MethodPost {
		// to do if a field was not given
		h.i.FallbackSearchResponse(w, h.c.FallbackTemplate)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = form.DecodeValues(&values, r.Form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if values.s == "" {
		// to do if a field was not given
		h.i.FallbackSearchResponse(w, h.c.FallbackTemplate)
		return
	}

	query := bleve.NewQueryStringQuery(values.s)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = values.pageSize
	searchRequest.From = values.pageSize + values.page

	results, err := h.i.Query(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = allTemplates.ExecuteTemplate(w, h.c.Template, results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Page is a standard data structure used to render markdown pages.
type Page struct {
	Title    string
	ToC      template.HTML
	Body     template.HTML
	Topics   []string
	Keywords []string
	Authors  []string
}

// Markdown is an http.Handler that renders a markdown file and serves it back.
//  Author and Topic tags before the first major title are parsed and displayed.
//  It is possible to restrict access to a page based on topic tag.
type Markdown struct {
	c ServerSection
}

func (h Markdown) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the request is empty, set it to the default.
	if r.URL.Path == "/" {
		r.URL.Path = path.Clean(h.c.Default)
	}

	// If the request doesn't end in .md, add that
	if path.Ext(r.URL.Path) != "md" {
		r.URL.Path = r.URL.Path + ".md"
	}

	pdata := new(PageMetadata)
	err := pdata.LoadPage(h.c.Path + r.URL.Path)
	if err != nil {
		log.Printf("request [ %s ] points bad file target [ %s ] sent to server",
			r.URL.Path, h.c.Path)
		http.Error(w, "Page not Found", http.StatusNotFound)
		return
	}

	if pdata.MatchedTopic(h.c.Restricted) {
		log.Printf("request [ %s ] was a page [ %s ] with a restricted tag",
			r.URL.Path, h.c.Path+r.URL.Path)
		http.Error(w, "Page not Found", http.StatusNotFound)
		//http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// parse any markdown in the input
	body := template.HTML(bodyParseMarkdown(pdata.Page))
	toc := template.HTML(tocParseMarkdown(pdata.Page))
	topics, keywords, authors := pdata.ListMeta()

	// ##TODO## put this template right in the function call
	// Then remove the Page Struct above
	response := Page{
		Title:    pdata.Title,
		ToC:      toc,
		Body:     body,
		Keywords: keywords,
		Topics:   topics,
		Authors:  authors,
	}
	err = allTemplates.ExecuteTemplate(w, h.c.Template, response)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// RawFile is a http.Handler that serves a raw file back, restricting by file
//  extension if necessary and adding appropiate mime-types.
type RawFile struct {
	c ServerSection
}

func (h RawFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the request is empty, set it to the default.
	if r.URL.Path == "/" {
		r.URL.Path = path.Clean(h.c.Default)
	}

	for _, restricted := range h.c.Restricted {
		if path.Ext(r.URL.Path) == restricted {
			log.Printf("request %s was has a disallowed extension %s",
				r.URL.Path, restricted)
			http.Error(w, "Request not allowed", 403)
			return
		}
	}

	f, err := os.Open(filepath.Join(h.c.Path, r.URL.Path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		log.Print(err)
		return
	}

	switch path.Ext(r.URL.Path) {
	case "js":
		w.Header().Set("Content-Type", "text/javascript")
	case "css":
		w.Header().Set("Content-Type", "text/css")
	case "gif":
		w.Header().Set("Content-Type", "image/gif")
	case "png":
		w.Header().Set("Content-Type", "image/png")
	case "jpg":
		w.Header().Set("Content-Type", "image/jpeg")
	case "jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	}

	_, err = io.Copy(w, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
	}
}

// FallbackSearchResponse is a function that writes a "bailout" template
func (i *Index) FallbackSearchResponse(w http.ResponseWriter,
	template string) {
	authors, err := i.ListField("author")
	if err != nil {
		http.Error(w, "failed to list authors", http.StatusInternalServerError)
		i.log.Println(err)
		return
	}
	topics, err := i.ListField("topic")
	if err != nil {
		http.Error(w, "failed to list topics", http.StatusInternalServerError)
		i.log.Println(err)
		return
	}

	fields := SearchResponse{Topics: topics, Authors: authors}

	err = allTemplates.ExecuteTemplate(w, template, fields)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		i.log.Println(err)
		return
	}
	return
}
