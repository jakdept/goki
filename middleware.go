package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
)

// BuildMuxer builds a handler by:
// * Opening Indexes
// * Adding all handlers under an index
// * Eventually putting all of the handlers together
func BuildMuxer(c GlobalSection, closer <-chan struct{},
	logs *log.Logger) (http.Handler, error) {
	m := http.NewServeMux()
	// ## TODO ## return an error instead of panic if overlapping muxes
	for _, r := range c.Redirects {
		m.Handle(r.Requested,
			http.RedirectHandler(r.Target, r.Code))
	}

	for _, i := range c.Indexes {
		var index Index
		var err error
		if i.IndexPath != "" {
			index, err = OpenIndex(i, logs)
			if err != nil {
				return nil, err
			}

			go func() {
				<-closer
				logs.Println(index.Close())
			}()
		}

		for _, h := range i.Handlers {
			switch h.ServerType {
			case "markdown":
				m.Handle(h.Prefix, http.StripPrefix(h.Prefix, Markdown{c: h}))
			case "raw":
				m.Handle(h.Prefix, http.StripPrefix(h.Prefix, RawFile{c: h}))
			case "query":
				m.Handle(h.Prefix, http.StripPrefix(h.Prefix, QueryHandler{c: h, i: index}))
			case "field":
				m.Handle(h.Prefix, http.StripPrefix(h.Prefix, FieldsHandler{c: h, i: index}))
			case "fuzzy":
				m.Handle(h.Prefix, http.StripPrefix(h.Prefix, FuzzyHandler{c: h, i: index}))
			}
		}
	}

	h := handlers.CompressHandler(m)
	return h, nil
}
