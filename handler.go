package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ajg/form"
	"github.com/blevesearch/bleve"
)

type fieldListHandler struct {
	config ServerSection
	index  *Index
}

func (h *fieldListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fields := strings.SplitN(r.URL.Path, "/", 2)

	if fields[0] == "" {
		// to do if a field was not given
		response, err := h.index.ListField(h.config.Default)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = allTemplates.ExecuteTemplate(w, h.config.Template,
			struct{ AllFields []string }{AllFields: response})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// to be done if a field was given
		response, err := h.index.ListFieldValues(h.config.Default, fields[0], 100, 1)
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

		err = allTemplates.ExecuteTemplate(w, h.config.Template, modifiedSearchResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type fuzzySearchHandler struct {
	config ServerSection
	index  *Index
}

func (h *fuzzySearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values := struct {
		s        string   `form:"s"`
		topics   []string `form:"topic"`
		authors  []string `form:"author"`
		page     int      `form:"page"`
		pageSize int      `form:"pageSize"`
	}{}

	err := form.Decode(&values, request.URL.Values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := bleve.NewQueryStringQuery(values.s)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = values.pageSize
	searchRequest.From = values.pageSize + values.page

	results, err := h.index.Query(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = allTemplates.ExecuteTemplate(w, h.config.Template, results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type querySearchHandler struct {
	config ServerSection
	index  *Index
}

func (h *querySearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values := struct {
		s        string `form:"s"`
		page     int    `form:"page"`
		pageSize int    `form:"pageSize"`
	}{}

	err := form.Decode(&values, request.URL.Values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := bleve.NewQueryStringQuery(values.s)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = values.pageSize
	searchRequest.From = values.pageSize + values.page

	results, err := h.index.Query(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = allTemplates.ExecuteTemplate(w, h.config.Template, results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type parsedHandler struct {
	config ServerSection
}

func (h *parsedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the request is empty, set it to the default.
	if requestPath == "/" {
		requestPath = path.Clean(h.serverConfig.Default)
	}

	// If the request doesn't end in .md, add that
	if path.Ext(requestPath) != "md" {
		requestPath = requestPath + ".md"
	}

	pdata := new(PageMetadata)
	err := pdata.LoadPage(serverConfig.Path + requestPath)
	if err != nil {
		log.Printf("request [ %s ] points to an bad file target [ %s ] sent to server %s",
			request.URL.Path, requestPath, serverConfig.Prefix)
		http.Error(responsePipe, "Page not Found", http.StatusNotFound)
		return
	}

	if pdata.MatchedTopic(serverConfig.Restricted) {
		log.Printf("request [ %s ] was against a page [ %s ] with a restricted tag",
			request.URL.Path, requestPath)
		http.Error(responsePipe, "Restricted Page", http.StatusNotFound)
		//http.Error(responsePipe, err.Error(), http.StatusForbidden)
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
	err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, response)
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
}

type rawHandler struct {
	config ServerSection
}

func (h *parsedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the request is empty, set it to the default.
	if requestPath == "/" {
		requestPath = path.Clean(h.serverConfig.Default)
	}

	for _, restricted := range h.config.Restricted {
		if path.Ext(requestPath) == restricted {
			log.Printf("request %s was improperly routed to the file handler with an disallowed extension %s", request.URL.Path, restricted)
			http.Error(responsePipe, "Request not allowed", 403)
			return
		}
	}

	f, err := os.Open(filePath.Join(h.config.Path, requestPath))
	if err != nil {
		http.Error(w, err.Error, http.StatusForbidden)
		log.Print(err)
		return
	}

	_, err = io.Copy(w, f)
	if err != nil {
		http.Error(w, err.Error, http.StatusInternalServerError)
		log.Print(err)
	}
}
