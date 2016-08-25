package main

import (
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
