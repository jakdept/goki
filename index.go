package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"

	"github.com/JackKnifed/blackfriday"
	"github.com/JackKnifed/blackfriday-text"
	"github.com/blevesearch/bleve"
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
	log        *log.Logger
	threads    sync.WaitGroup
	closer     chan struct{}
}

func OpenIndex(c IndexSection, l *log.Logger) (*Index, error) {
	i := &Index{config: c, log: l}
	index, err := bleve.Open(path.Clean(i.config.IndexPath))
	if err == nil {
		i.index = index
	} else if err == bleve.ErrorIndexPathDoesNotExist {
		indexMapping := i.buildIndexMapping()
		index, err = bleve.New(path.Clean(i.config.IndexPath), indexMapping)
		if err != nil {
			return &Index{}, &Error{Code: ErrIndexCreate, value: c.IndexPath, innerError: err}
		}

		i.index = index
	}

	// i am confused how the filepath -> uriPaths line up
	// start watchers
	for each, _ := range i.config.WatchDirs {
		path := filepath.Clean(each)
		go i.WatchDir(path)
		go i.CrawlDir(path)
	}

	// prime the closing channel
	go func() {
		<-i.closer
	}()
	i.closer <- struct{}{}

	return i, nil
}

func (i *Index) buildIndexMapping() *bleve.IndexMapping {

	// create a text field type
	enTextFieldMapping := bleve.NewTextFieldMapping()
	enTextFieldMapping.Analyzer = i.config.IndexType

	// create a date field type
	dateTimeMapping := bleve.NewDateTimeFieldMapping()

	// map out the wiki page
	wikiMapping := bleve.NewDocumentMapping()
	wikiMapping.AddFieldMappingsAt("title", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("path", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("body", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("topic", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("keyword", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("author", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("modified", dateTimeMapping)

	// add the wiki page mapping to a new index
	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping(i.config.IndexName, wikiMapping)
	indexMapping.DefaultAnalyzer = i.config.IndexType

	return indexMapping
}

func (i *Index) Close() error {
	// dump something into the channel incase it's currently nil
	i.closer <- struct{}{}
	close(i.closer)

	i.threads.Wait()

	if err := i.index.Close(); err != nil {
		return &Error{Code: ErrIndexClose, innerError: err}
	}
	return nil
}

func (i *Index) Wipe() error {
	i.lock.Lock()
	if err := i.index.Close(); err != nil {
		i.lock.Unlock()
		return &Error{Code: ErrIndexClose, innerError: err}
	}

	if err := os.RemoveAll(i.config.IndexPath); err != nil {
		return &Error{Code: ErrIndexRemove, value: i.config.IndexPath, innerError: err}
	}

	// TODO not sure if needed
	if err := os.Mkdir(i.config.IndexPath, 644); err != nil {
		return &Error{Code: ErrIndexCreate, value: i.config, innerError: err}
	}
	// remove index

	indexMapping := i.buildIndexMapping()
	index, err := bleve.New(path.Clean(i.config.IndexPath), indexMapping)
	if err != nil {
		close(i.closer)
		return &Error{Code: ErrIndexCreate, value: i.config, innerError: err}
	}

	i.index = index
	i.lock.Unlock()
	return nil
}

func (i *Index) WatchDir(watchPath string) error {
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

func (i *Index) indexFileFunc(rootPath string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			i.log.Println(i.UpdateURI(path, i.getURI(path, rootPath)))
		}
		return nil
	}
}

func (i *Index) CrawlDir(path string) {
	filepath.Walk(path, i.indexFileFunc(path))
}

func (i *Index) DeleteURI(uriPath string) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.log.Printf("removing %s", uriPath)
	err := i.index.Delete(uriPath)
	if err != nil {
		return &Error{Code: ErrIndexError, value: uriPath, innerError: err}
	}
	return nil
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

func (i *Index) Query(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()

	searchResults, err := i.index.Search(request)
	if err != nil {
		return &bleve.SearchResult{}, &Error{Code: ErrInvalidQuery, innerError: err}
	}

	return searchResults, nil
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

func (i *Index) getURI(filePath, filePrefix string) (uriPath string) {
	uriPath = strings.TrimPrefix(filePath, filePrefix)
	uriPath = strings.TrimPrefix(uriPath, "/")
	uriPrefix := strings.TrimSuffix(i.config.IndexPath, "/")
	uriPath = uriPrefix + "/" + uriPath
	return
}
