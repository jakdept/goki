package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	blevemapping "github.com/blevesearch/bleve/mapping"

	"github.com/JackKnifed/blackfriday"
	blackfridaytext "github.com/JackKnifed/blackfriday-text"
	fsnotify "gopkg.in/fsnotify.v1"
)

type indexedPage struct {
	Title    string    `json:"title"`
	URIPath  string    `json:"path"`
	Body     string    `json:"body"`
	Topics   string    `json:"topic"`
	Keywords string    `json:"keyword"`
	Authors  string    `json:"author"`
	Modified time.Time `json:"modified"`
}

type Index interface {
	Close() error
	Wipe() error
	CrawlDir(string, string) error
	WatchDir(string, string) error
	Query(*bleve.SearchRequest) (*bleve.SearchResult, error)
	// CreateResponseData(*bleve.SearchResult, int) (SearchResponse, error)
	// ListField(string) ([]string, error)
	// ListAllField(string, string, int, int) (SearchResponse, error)
	// FuzzySearch(FuzzySearchValues) (SearchResponse, error)
	// QuerySearch(string, int, int) (SearchResponse, error)
	// FallbackSearchResponse(http.ResponseWriter, string)
}

type indexObject struct {
	index      bleve.Index
	lock       sync.RWMutex
	updateChan chan indexedPage
	config     IndexSection
	log        *log.Logger
	threads    sync.WaitGroup
	closer     chan struct{}
}

func OpenIndex(c IndexSection, l *log.Logger) (Index, error) {
	i := &indexObject{config: c, log: l}
	index, err := bleve.Open(path.Clean(i.config.IndexPath))
	if err == nil {
		i.index = index
	} else if err == bleve.ErrorIndexPathDoesNotExist {
		indexMapping := i.buildIndexMapping()
		index, err = bleve.New(path.Clean(i.config.IndexPath), indexMapping)
		if err != nil {
			return nil, &Error{Code: ErrIndexCreate, value: c.IndexPath, innerError: err}
		}

		i.index = index
	}

	// start watchers
	for filePrefix, uriPrefix := range i.config.WatchDirs {
		filePrefix := filepath.Clean(filePrefix)
		uriPrefix = path.Clean(uriPrefix)
		if !strings.HasSuffix(filePrefix, "/") {
			filePrefix += "/"
		}
		if !strings.HasSuffix(uriPrefix, "/") {
			uriPrefix += "/"
		}

		go i.WatchDir(filePrefix, uriPrefix)
		go i.CrawlDir(filePrefix, uriPrefix)
		i.log.Printf("watching and walking [%s]", filePrefix)
	}

	// prime the closing channel
	i.closer = make(chan struct{}, 1)
	i.closer <- struct{}{}
	<-i.closer

	return i, nil
}

func (i *indexObject) buildIndexMapping() blevemapping.IndexMapping {

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

func (i *indexObject) Close() error {
	// dump something into the channel incase it's currently nil
	i.closer <- struct{}{}
	close(i.closer)

	i.threads.Wait()

	if err := i.index.Close(); err != nil {
		return &Error{Code: ErrIndexClose, innerError: err}
	}
	return nil
}

func (i *indexObject) Wipe() error {
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

func (i *indexObject) WatchDir(watchPath, uriPrefix string) error {
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
			i.log.Fatal(err)
		case <-idleTimer.C:
			for _, event := range queuedEvents {
				if filepath.Ext(event.Name) == i.config.WatchExtension {
					switch event.Op {
					case fsnotify.Remove, fsnotify.Rename:
						// delete the filePath
						i.DeleteURI(uriPrefix + event.Name)
					case fsnotify.Create, fsnotify.Write:
						// update the filePath
						i.UpdateURI(watchPath+event.Name, uriPrefix+event.Name)
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

func (i *indexObject) indexFileFunc(rootPath, uriPrefix string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			err := i.UpdateURI(path, i.getURI(path, rootPath, uriPrefix))
			if err != nil {
				i.log.Println(err)
			}
		}
		return nil
	}
}

func (i *indexObject) CrawlDir(path, uriPrefix string) error {
	err := filepath.Walk(path, i.indexFileFunc(path, uriPrefix))
	if err != nil {
		return err
	}
	return nil
}

func (i *indexObject) DeleteURI(uriPath string) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.log.Printf("removing %s", uriPath)
	err := i.index.Delete(uriPath)
	if err != nil {
		return &Error{Code: ErrIndexError, value: uriPath, innerError: err}
	}
	return nil
}

func (i *indexObject) UpdateURI(filePath, uriPath string) error {
	page, err := i.generateWikiFromFile(filePath, uriPath)
	if err != nil {
		return err
	}

	i.lock.Lock()
	defer i.lock.Unlock()
	i.log.Printf("updated: [%s] as [%s] indexed at [%s]", filePath, page.URIPath, uriPath)
	err = i.index.Index(uriPath, page)
	if err != nil {
		return &Error{Code: ErrIndexError, value: uriPath, innerError: err}
	}
	return nil
}

func (i *indexObject) Query(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()

	// dump, _ := blevequery.DumpQuery(i.index.Mapping(), request.Query)
	// log.Println(dump)

	searchResults, err := i.index.Search(request)
	if err != nil {
		return &bleve.SearchResult{}, &Error{Code: ErrInvalidQuery, innerError: err}
	}

	// log.Printf("\nresults: %#v\n", searchResults.Hits)
	return searchResults, nil
}

func (i *indexObject) generateWikiFromFile(filePath, uriPath string) (*indexedPage, error) {
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
		URIPath:  strings.TrimSuffix(uriPath, ".md"),
		Topics:   strings.Join(topics, " "),
		Keywords: strings.Join(keywords, " "),
		Authors:  strings.Join(authors, " "),
		Modified: pdata.FileStats.ModTime(),
	}

	return &rv, nil
}

func (i *indexObject) cleanupMarkdown(input []byte) string {
	extensions := 0 | blackfriday.EXTENSION_ALERT_BOXES
	renderer := blackfridaytext.TextRenderer()
	output := blackfriday.Markdown(input, renderer, extensions)
	return string(output)
}

func (i *indexObject) getURI(filePath, trimPrefix, addPrefix string) string {
	return addPrefix + strings.TrimPrefix(filePath, trimPrefix)
}
