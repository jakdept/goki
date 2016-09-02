package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
)

/*
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
*/

/*
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
*/

func BuildMuxer(c GlobalSection, closer <-chan struct{},
	logs *log.Logger) (http.Handler, error) {
	m := http.NewServeMux()
	// ## TODO ## return an error instead of panic if overlapping muxes
	for _, i := range c.Indexes {
		if i.IndexPath != "" {
			index, err := OpenIndex(i, logs)
			if err != nil {
				return nil, err
			}

			go func() {
				<-closer
				logs.Println(index.Close())
			}()

			for _, h := range i.Handlers {
				switch h.ServerType {
				case "markdown":
					m.Handle(h.Prefix, http.StripPrefix(h.Prefix, Markdown{c: h}))
				case "raw":
					m.Handle(h.Prefix, http.StripPrefix(h.Prefix, RawFile{c: h}))
				case "query":
					m.Handle(h.Prefix, http.StripPrefix(h.Prefix, QuerySearch{c: h, i: index}))
				case "field":
					m.Handle(h.Prefix, http.StripPrefix(h.Prefix, Fields{c: h, i: index}))
				case "fuzzy":
					m.Handle(h.Prefix, http.StripPrefix(h.Prefix, FuzzySearch{c: h, i: index}))
				}
			}
			for _, r := range i.Redirects {
				m.Handle(r.Requested,
					http.RedirectHandler(r.Target, r.Code))
			}
		}
	}

	h := handlers.CompressHandler(m)
	return h, nil
}
