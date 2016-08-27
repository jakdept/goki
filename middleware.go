package main

import (
	"log"
	"net/http"
	"path"
	"strings"
)

type trimPrefixHandler struct {
	prefix  string
	handler http.Handler
}

func (h *trimPrefixHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, path.Clean(h.prefix)) {
		http.Error(w, "incorrect path for handler", http.StatusInternalServerError)
	}
	r.URL.Path = path.Clean(strings.TrimPrefix(r.URL.Path, path.Clean(h.prefix)))
	h.handler(w, r)
}

// this is the old way of setting up things #TODO#
func MakeHandler(handlerConfig ServerSection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToLower(handlerConfig.ServerType) {
		case "markdown":
			MarkdownHandler(w, r, handlerConfig)
		case "raw":
			RawHandler(w, r, handlerConfig)
		case "querysearch":
			log.Printf("serving a Query String Search on [%s] with the template [%s]", handlerConfig.Prefix, handlerConfig.Template)
			QuerySearchHandler(w, r, handlerConfig)
		case "fieldlist":
			FieldListHandler(w, r, handlerConfig)
		case "search":
			log.Printf("serving a Search on [%s] with the template [%s]", handlerConfig.Prefix, handlerConfig.Template)
			SearchHandler(w, r, handlerConfig)
		default:
			log.Printf("Bad server type [%s]", handlerConfig.ServerType)
		}
	}
}
