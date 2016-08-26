package main

import (
	"github.com/ajg/form"
	"net/http"
	"strings"
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
		s string `form:"s"`
		topics []string `form:"topic"`
		authors []string `form:"author"`
		page int `form:"page"`
		pageSize int `form:"pageSize"`
	}

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

	
	results, err  := h.index.Query(searchRequest)
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
		s string `form:"s"`
		page int `form:"page"`
		pageSize int `form:"pageSize"`
	}

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

	
	results, err  := h.index.Query(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = allTemplates.ExecuteTemplate(w, h.config.Template, results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
