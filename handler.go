package main

import (
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
