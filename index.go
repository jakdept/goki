package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"

	"github.com/JackKnifed/blackfriday"
	"github.com/blevesearch/bleve"
	"github.com/mschoch/blackfriday-text"
)

type indexedPage struct {
	Title    string    `json:"title"`
	URIPath  string    `json:"path"`
	Body     string    `json:"body"`
	Topics   string    `json:"topic"`
	Keywords string    `json:"keyword"`
	Authors  string    `json: "author"`
	Modified time.Time `json:"modified"`
}

type Index struct {
	index      bleve.Index
	lock       sync.RWMutex
	updateChan chan indexedPage
	config     IndexSection
	log        log.Logger
	threads    sync.WaitGroup
	closer     chan struct{}
}

func OpenIndex(config IndexSection) (*Index, error) {

}

func (i *Index) WatchDir(watchPath string) error {
	if !strings.HasSuffix(watchPath, "/") {
		watchPath = watchPath + "/"
	}

	i.log.Printf("watching '%s' for changes...", watchPath)

	i.threads.Add(1)
	defer i.threads.Done()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return &Error{Code: ErrWatcherCreate, innerError: err}
	}
	defer watcher.Close()

	err = watcher.Add(strings.TrimSuffix(watchPath, "/"))
	if err != nil {
		return &Error{Code: ErrWatcherAdd, value: watchPath, innerError: err}
	}

	idleTimer := time.NewTimer(10 * time.Second)
	var queuedEvents []fsnotify.Event
	var more bool

	for more {
		select {
		case _, more = <-i.closer:
		case event := <-watcher.Events:
			queuedEvents = append(queuedEvents, event)
			idleTimer.Reset(10 * time.Second)
		case err := <-watcher.Errors:
			log.Fatal(err)
		case <-idleTimer.C:
			for _, event := range queuedEvents {
				if strings.HasSuffix(event.Name, i.config.WatchExtension) {
					switch event.Op {
					case fsnotify.Remove, fsnotify.Rename:
						// delete the filePath
						i.DeleteURI(i.config.IndexPath + event.Name)
					case fsnotify.Create, fsnotify.Write:
						// update the filePath
						i.UpdateURI(watchPath+event.Name, i.config.IndexPath+event.Name)
					default:
						// no changes, so repeat the cycle
					}
				}
			}
			queuedEvents = make([]fsnotify.Event, 0)
			idleTimer.Reset(10 * time.Second)
		}
	}
	return nil
}

func (i *Index) CrawlDir(globPath string) error {

}

func (i *Index) DeleteURI(uriPath string) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.log.Printf("removing %s", uriPath)
	err := i.index.Delete(uriPath)
	if err != nil {
		return &Error{Code: ErrIndexError, value: uriPath, innerError: err}
	}
}

func (i *Index) UpdateURI(filePath, uriPath string) error {
	page, err := i.generateWikiFromFile(filePath, uriPath)
	if err != nil {
		return err
	}

	i.lock.Lock()
	defer i.lock.Unlock()
	i.log.Printf("updated: %s as %s", filePath, uriPath)
	err = i.index.Index(uriPath, page)
	if err != nil {
		return &Error{Code: ErrIndexError, value: uriPath, innerError: err}
	}
	return nil
}

func (i *Index) Query(bleve.SearchRequest) (bleve.SearchResponse, error) {

}

func (i *Index) ListField(field string) ([]string, error) {

}

func (i *Index) generateWikiFromFile(filePath, uriPath string) (*indexedPage, error) {
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filePath)
	if err != nil {
		return nil, err
	}

	if pdata.MatchedTopic(i.config.Restricted) == true {
		return nil, &Error{Code: ErrPageRestricted, value: pdata.Title}
	}

	topics, keywords, authors := pdata.ListMeta()
	rv := indexedPage{
		Title:    pdata.Title,
		Body:     i.cleanupMarkdown(pdata.Page),
		URIPath:  uriPath,
		Topics:   strings.Join(topics, " "),
		Keywords: strings.Join(keywords, " "),
		Authors:  strings.Join(authors, " "),
		Modified: pdata.FileStats.ModTime(),
	}

	return &rv, nil
}

func (i *Index) cleanupMarkdown(input []byte) string {
	extensions := 0 | blackfriday.EXTENSION_ALERT_BOXES
	renderer := blackfridaytext.TextRenderer()
	output := blackfriday.Markdown(input, renderer, extensions)
	return string(output)
}

func (i *Index) getURI(filePath, filePrefix) (uriPath string) {
	uriPath = strings.TrimPrefix(filePath, filePrefix)
	uriPath = strings.TrimPrefix(uriPath, "/")
	uriPrefix = strings.TrimSuffix(uriPrefix, "/")
	uriPath = uriPrefix + "/" + uriPath
}
