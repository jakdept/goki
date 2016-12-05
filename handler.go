package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/ajg/form"
)

// FieldsHandler is a standard handler that pulls the first folder of the
//  response, and lists that topic or author. If there is none, it falls back to
//  listing all topics or authors with the fallback template.
type FieldsHandler struct {
	c ServerSection
	i Index
}

func (h FieldsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// fields := strings.SplitN(r.URL.Path, "/", 2)

	// if len(fields) < 2 || fields[1] == "" {
	if r.URL.Path == "" {
		// to do if a field was not given
		switch h.c.FallbackTemplate {
		case "":
			FallbackSearchResponse(h.i, w, h.c.Template)
		default:
			FallbackSearchResponse(h.i, w, h.c.FallbackTemplate)
		}
		return
	}

	// to be done if a field was given - might actually have to be 1 idk
	// results, err := ListAllField(h.i, h.c.Default, fields[0], 100, 1)
	results, err := ListAllField(h.i, h.c.Default, r.URL.Path, 100, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = allTemplates.ExecuteTemplate(w, h.c.Template, results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// FuzzySearch is a normal search format - it should provide a point and click interface to allow searching.
type FuzzyHandler struct {
	c ServerSection
	i Index
}

func (h FuzzyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var values FuzzySearchValues

	if _, ok := r.Form["s"]; ok && len(r.Form["s"]) > 0 {
		values.Term = r.Form["s"][0]
	}
	if _, ok := r.Form["topic"]; ok && len(r.Form["topic"]) > 0 {
		values.Topics = r.Form["topic"]
	}
	if _, ok := r.Form["author"]; ok && len(r.Form["author"]) > 0 {
		values.Authors = r.Form["author"]
	}

	if _, ok := r.Form["page"]; ok && len(r.Form["page"]) > 0 {
		i, err := strconv.Atoi(r.Form["page"][0])
		if err == nil {
			values.Page = i
		}
	}
	if _, ok := r.Form["pageSize"]; ok && len(r.Form["pageSize"]) > 0 {
		i, err := strconv.Atoi(r.Form["pageSize"][0])
		if err == nil {
			values.Page = i
		}
	}

	results, err := FuzzySearch(h.i, values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("\n%#v\n%#v\n", values, results)

	err = allTemplates.ExecuteTemplate(w, h.c.Template, results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// QuerySearchHAndler is a handler that uses a custom search format to do
// custom queries.
type QueryHandler struct {
	c ServerSection
	i Index
}

func (h QueryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values := struct {
		s        string `form:"s"`
		page     int    `form:"page"`
		pageSize int    `form:"pageSize"`
	}{}

	if r.Method != http.MethodPost {
		// to do if a field was not given
		FallbackSearchResponse(h.i, w, h.c.FallbackTemplate)
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
		FallbackSearchResponse(h.i, w, h.c.FallbackTemplate)
		return
	}

	results, err := QuerySearch(h.i, values.s, values.page, values.pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = allTemplates.ExecuteTemplate(w, h.c.Template, results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RawFile is a http.Handler that serves a raw file back, restricting by file
//  extension if necessary and adding approipate mime-types.
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
		w.Header().Set("Content-Type", "application/javascript")
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
	defaultPage(h.c.Default, appendExtension(".md", h.backendServe())).ServeHTTP(w, r)
}

func (h Markdown) backendServe() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pdata := new(PageMetadata)
		filePath := filepath.Join(h.c.Path, r.URL.Path)
		err := pdata.LoadPage(filePath)
		if err != nil {
			log.Printf("request [ %s ] points bad file target [ %s ] sent to server - %v",
				r.URL.Path, filePath, err)
			http.Error(w, "Page not Found", http.StatusNotFound)
			return
		}

		if pdata.MatchedTopic(h.c.Restricted) {
			log.Printf("request [ %s ] was a page [ %s ] with a restricted tag",
				r.URL.Path, filePath)
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
	})
}

func appendExtension(suffix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request doesn't end in .md, add that
		if path.Ext(r.URL.Path) != suffix {
			r.URL.Path = r.URL.Path + suffix
		}
		h.ServeHTTP(w, r)
	})
}

func defaultPage(page string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request is empty, set it to the default.
		r.URL.Path = path.Clean(r.URL.Path)
		if r.URL.Path == "." {
			r.URL.Path = path.Clean(page)
		}
		h.ServeHTTP(w, r)
	})
}
