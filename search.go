package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/JackKnifed/blackfriday"
	"github.com/JackKnifed/blackfriday-text"
	"github.com/blevesearch/bleve"
	"gopkg.in/fsnotify.v1"
)

var openWatchers []fsnotify.Watcher

func createIndex(config IndexSection) bool {
	newIndex, err := bleve.Open(path.Clean(config.IndexPath))
	if err == nil {
		log.Printf("Index already exists %s", config.IndexPath)
	} else if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Creating new index %s", config.IndexName)
		// create a mapping
		indexMapping := buildIndexMapping(config)
		newIndex, err = bleve.New(path.Clean(config.IndexPath), indexMapping)
		if err != nil {
			log.Printf("Failed to create the new index %s - %v", config.IndexPath, err)
			return false
		} else {
		}
	} else {
		log.Printf("Got an error opening the index %s but it already exists %v", config.IndexPath, err)
		return false
	}
	newIndex.Close()
	return true
}

func EnableIndex(config IndexSection) bool {
	if !createIndex(config) {
		return false
	}
	for dir, path := range config.WatchDirs {
		// dir = strings.TrimSuffix(dir, "/")
		log.Printf("Watching and walking dir %s index %s", dir, config.IndexPath)
		walkForIndexing(dir, dir, path, config)
	}
	return true
}

func DisableAllIndexes() {
	log.Println("Stopping all watchers")
	for _, watcher := range openWatchers {
		watcher.Close()
	}
}

func buildIndexMapping(config IndexSection) *bleve.IndexMapping {

	// create a text field type
	enTextFieldMapping := bleve.NewTextFieldMapping()
	enTextFieldMapping.Analyzer = config.IndexType

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
	indexMapping.AddDocumentMapping(config.IndexName, wikiMapping)
	indexMapping.DefaultAnalyzer = config.IndexType

	return indexMapping
}

// walks a given path, and runs processUpdate on each File
func walkForIndexing(path, filePath, requestPath string, config IndexSection) {
	//watcherLoop(path, filePath, requestPath, config)
	dirEntries, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, dirEntry := range dirEntries {
		dirEntryPath := path + string(os.PathSeparator) + dirEntry.Name()
		if dirEntry.IsDir() {
			walkForIndexing(dirEntryPath, filePath, requestPath, config)
		} else if strings.HasSuffix(dirEntry.Name(), config.WatchExtension) {
			processUpdate(dirEntryPath, getURIPath(dirEntryPath, filePath, requestPath), config)
		}
	}
}

// given all of the inputs, watches afor new/deleted files in that directory
// adds/removes/udpates index as necessary
func watcherLoop(watchPath, filePrefix, uriPrefix string, config IndexSection) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add(watchPath)
	if err != nil {
		log.Fatal(err)
	}

	idleTimer := time.NewTimer(10 * time.Second)
	queuedEvents := make([]fsnotify.Event, 0)

	openWatchers = append(openWatchers, *watcher)

	log.Printf("watching '%s' for changes...", watchPath)

	for {
		select {
		case event := <-watcher.Events:
			queuedEvents = append(queuedEvents, event)
			idleTimer.Reset(10 * time.Second)
		case err := <-watcher.Errors:
			log.Fatal(err)
		case <-idleTimer.C:
			for _, event := range queuedEvents {
				if strings.HasSuffix(event.Name, config.WatchExtension) {
					switch event.Op {
					case fsnotify.Remove, fsnotify.Rename:
						// delete the filePath
						processDelete(getURIPath(watchPath+event.Name, filePrefix, uriPrefix),
							config.IndexName)
					case fsnotify.Create, fsnotify.Write:
						// update the filePath
						processUpdate(watchPath+event.Name,
							getURIPath(watchPath+event.Name, filePrefix, uriPrefix), config)
					default:
						// ignore
					}
				}
			}
			queuedEvents = make([]fsnotify.Event, 0)
			idleTimer.Reset(10 * time.Second)
		}
	}
}

// Update the entry in the index to the output from a given file
func processUpdate(filePath, uriPath string, config IndexSection) {
	page, err := generateWikiFromFile(filePath, uriPath, config.Restricted)
	if err != nil {
		log.Print(err)
	} else {
		index, _ := bleve.Open(config.IndexPath)
		defer index.Close()
		index.Index(uriPath, page)
		log.Printf("updated: %s as %s", filePath, uriPath)
	}
}

// Deletes a given path from the wiki entry
func processDelete(uriPath, indexPath string) {
	log.Printf("delete: %s", uriPath)
	index, _ := bleve.Open(indexPath)
	defer index.Close()
	err := index.Delete(uriPath)
	if err != nil {
		log.Print(err)
	}
}

func cleanupMarkdown(input []byte) string {
	extensions := 0 | blackfriday.EXTENSION_ALERT_BOXES
	renderer := blackfridaytext.TextRenderer()
	output := blackfriday.Markdown(input, renderer, extensions)
	return string(output)
}

func generateWikiFromFile(filePath, uriPath string, restrictedTopics []string) (*indexedPage, error) {
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filePath)
	if err != nil {
		return nil, err
	}

	if pdata.MatchedTopic(restrictedTopics) == true {
		return nil, errors.New("Hit a restricted page - " + pdata.Title)
	}

	topics, keywords, authors := pdata.ListMeta()
	rv := indexedPage{
		Title:    pdata.Title,
		Body:     cleanupMarkdown(pdata.Page),
		URIPath:  uriPath,
		Topics:   strings.Join(topics, " "),
		Keywords: strings.Join(keywords, " "),
		Authors:  strings.Join(authors, " "),
		Modified: pdata.FileStats.ModTime(),
	}

	return &rv, nil
}

func getURIPath(filePath, filePrefix, uriPrefix string) (uriPath string) {
	uriPath = strings.TrimPrefix(filePath, filePrefix)
	uriPath = strings.TrimPrefix(uriPath, "/")
	uriPrefix = strings.TrimSuffix(uriPrefix, "/")
	uriPath = uriPrefix + "/" + uriPath
	return
}

// given a path to an index, and a name of field to check
// lists all unique values for that field in index
func ListField(indexPath, field string) ([]string, error) {
	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)

	facet := bleve.NewFacetRequest(field, 1)
	searchRequest.AddFacet("allValues", facet)
	if err := query.Validate(); err != nil {
		return nil, err
	}

	// Open the index
	index, err := bleve.Open(indexPath)
	defer index.Close()
	if index == nil {
		return nil, fmt.Errorf("no such index [%s]", indexPath)
	} else if err != nil {
		return nil, fmt.Errorf("problem opening index [%s] - %s", indexPath, err)
	}

	searchResults, err := index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	results := *new([]string)
	for _, oneTerm := range searchResults.Facets["allValues"].Terms {
		results = append(results, oneTerm.Term)
	}

	return results, nil
}

type SearchResponse struct {
	// AllFields []string
	TotalHits  int
	PageOffset int
	SearchTime time.Duration
	Topics     []string
	Authors    []string
	Results    []SearchResponseResult
}

type SearchResponseResult struct {
	Title    string
	URIPath  string
	Score    float64
	Topics   []string
	Keywords []string
	Authors  []string
	Body     string
}

func CreateResponseData(rawResults bleve.SearchResult, pageOffset int, topics []string, authors []string) (SearchResponse, error) {
	var response SearchResponse

	response.TotalHits = int(rawResults.Total)
	response.PageOffset = pageOffset
	response.SearchTime = rawResults.Took
	response.Topics = topics
	response.Authors = authors
	for _, hit := range rawResults.Hits {
		var newHit SearchResponseResult

		newHit.Score = float64(hit.Score * 100 / rawResults.MaxScore)

		if _, isThere := hit.Fields["title"]; !isThere {
			newHit.Title = ""
		} else if str, ok := hit.Fields["title"].(string); ok {
			newHit.Title = str
		} else {
			return response, errors.New("returned title was not a string")
		}

		if _, isThere := hit.Fields["path"]; !isThere {
			newHit.URIPath = ""
		} else if str, ok := hit.Fields["path"].(string); ok {
			newHit.URIPath = str
		} else {
			return response, errors.New("returned path was not a string")
		}

		if _, isThere := hit.Fields["body"]; !isThere {
			newHit.Body = ""
		} else if str, ok := hit.Fields["body"].(string); ok {
			newHit.Body = str
		} else {
			return response, errors.New("returned body was not a string")
		}

		if _, isThere := hit.Fields["topic"]; !isThere {
			newHit.Topics = []string{}
		} else if str, ok := hit.Fields["topic"].(string); ok {
			newHit.Topics = strings.Split(str, " ")
		} else {
			return response, errors.New("returned topics were not a string")
		}

		if _, isThere := hit.Fields["keyword"]; !isThere {
			newHit.Keywords = []string{}
		} else if str, ok := hit.Fields["keyword"].(string); ok {
			newHit.Keywords = strings.Split(str, " ")
		} else {
			return response, errors.New("returned keywords were not a string")
		}

		if _, isThere := hit.Fields["author"]; !isThere {
			newHit.Authors = []string{}
		} else if str, ok := hit.Fields["author"].(string); ok {
			newHit.Authors = strings.Split(str, " ")
		} else {
			return response, errors.New("returned authors were not a string")
		}

		response.Results = append(response.Results, newHit)
	}

	return response, nil
}