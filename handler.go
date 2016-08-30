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

type Fields struct {
	c ServerSection
	i *Index
}

func (h Fields) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fields := strings.SplitN(r.URL.Path, "/", 2)

	if fields[0] == "" {
		// to do if a field was not given
		response, err := h.i.ListField(h.c.Default)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = allTemplates.ExecuteTemplate(w, h.c.Template,
			struct{ AllFields []string }{AllFields: response})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

type FuzzySearch struct {
	c ServerSection
	i *Index
}

func (h FuzzySearch) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values := struct {
		s        string   `form:"s"`
		topics   []string `form:"topic"`
		authors  []string `form:"author"`
		page     int      `form:"page"`
		pageSize int      `form:"pageSize"`
	}{}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = form.DecodeValues(&values, r.Form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

type Page struct {
	Title    string
	ToC      template.HTML
	Body     template.HTML
	Topics   []string
	Keywords []string
	Authors  []string
}

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

	_, err = io.Copy(w, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
	}
}

func (i *Index) FallbackSearchResponse(w http.ResponseWriter,
	template string) error {
	authors, err := i.ListField("author")
	if err != nil {
		http.Error(w, "failed to list authors", http.StatusInternalServerError)
		return err
	}
	topics, err := i.ListField("topic")
	if err != nil {
		http.Error(w, "failed to list topics", http.StatusInternalServerError)
		return err
	}

	fields := SearchResponse{Topics: topics, Authors: authors}

	err = allTemplates.ExecuteTemplate(w, template, fields)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}
