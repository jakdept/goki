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
)

// FieldsHandler is a standard handler that pulls the first folder of the
//  response, and lists that topic or author. If there is none, it falls back to
//  listing all topics or authors with the fallback template.
type FieldsHandler struct {
	c ServerSection
	i Index
}

func (h FieldsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fields := strings.SplitN(r.URL.Path, "/", 2)

	if len(fields) < 2 || fields[1] == "" {
		// to do if a field was not given
		FallbackSearchResponse(h.i, w, h.c.FallbackTemplate)
		return
	}
	// to be done if a field was given - might actually have to be 1 idk
	results, err := ListAllField(h.i, h.c.Default, fields[0], 100, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = allTemplates.ExecuteTemplate(w, h.c.Template, results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// FuzzySearchHandler is a normal search format - it should provide a point and
//  click interface to allow searching.
type FuzzyHandler struct {
	c ServerSection
	i Index
}

func (h FuzzyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	var values FuzzySearchValues
	err = form.DecodeValues(&values, r.Form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := FuzzySearch(h.i, values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	r.URL.Path = path.Clean(r.URL.Path)
	if r.URL.Path == "." {
		r.URL.Path = path.Clean(h.c.Default)
	}

	// If the request doesn't end in .md, add that
	if path.Ext(r.URL.Path) != "md" {
		r.URL.Path = r.URL.Path + ".md"
	}

	pdata := new(PageMetadata)
	filePath := filepath.Join(h.c.Path, r.URL.Path)
	err := pdata.LoadPage(filePath)
	if err != nil {
		log.Printf("request [ %s ] points bad file target [ %s ] sent to server",
			r.URL.Path, filePath)
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
