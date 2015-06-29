package gnosis

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/JackKnifed/blackfriday"
	"github.com/blevesearch/bleve"
	"github.com/mschoch/blackfriday-text"
	"gopkg.in/fsnotify.v1"
)

type GnosisIndex struct {
	TrueIndex bleve.Index
	Config    IndexSection
	// ##TODO## figure this out
	// I cannot use append and range on an array of pointers
	// and yet I do not think I can modify an array of non-pointers
	// IDK what to do just yet
	openWatchers []fsnotify.Watcher
}

type WatcherMeta struct {
	watcher fsnotify.Watcher
	filePath string
	requestPath string
	indexPath string
}

var openWatchers []WatcherMeta

type indexedPage struct {
	Name     string    `json:"name"`
	Filepath string    `json:"path"`
	Body     string    `json:"body"`
	Topics   string    `json:"topic"`
	Keywords string    `json:"keyword"`
	Modified time.Time `json:"modified"`
}

func CreateIndex(config IndexSection) error {
	newIndex, err := bleve.Open(path.Clean(IndexSection.IndexPath))
	if err == nil {
		log.Printf("Index already exists %s", IndexSection.IndexPath)
		defer newIndex.Close()
	} else if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Creating new index %s", IndexSection.IndexName)
		// create a mapping
		indexMapping := buildIndexMapping(IndexSection)
		newIndex, err := bleve.New(path.Clean(IndexSection.IndexPath), indexMapping)
		if err != nil {
			log.Printf("Failed to create the new index %s - %v", IndexSection.IndexPath, err)
		} else {
		defer newIndex.Close()
		}
	} else {
		log.Printf("Got an error opening the index %s but it already exists %v", IndexSection.IndexPath, err)
	}
}

func EnableIndex(config IndexSection) {
	for dir, path := range config.WatchDirs {
		// dir = strings.TrimSuffix(dir, "/")
		log.Printf("Watching and walking dir %s index %s", dir, config.IndexPath)
		watcher := startWatching(dir, path, config)
		openWatchers = append(openWatchers, watcher)
		walkForIndexing(dir, dir, path, config.IndexPath)
	}
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
	wikiMapping.AddFieldMappingsAt("path", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("body", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("topic", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("keyword", enTextFieldMapping)
	wikiMapping.AddFieldMappingsAt("modified", dateTimeMapping)

	// add the wiki page mapping to a new index
	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping(config.IndexName, wikiMapping)
	indexMapping.DefaultAnalyzer = config.IndexType

	return indexMapping
}

// walks a given path, and runs processUpdate on each File
func walkForIndexing(path string, origPath string, requestPath string, config IndexSection) {
	dirEntries, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, dirEntry := range dirEntries {
		dirEntryPath := path + string(os.PathSeparator) + dirEntry.Name()
		if dirEntry.IsDir() {
			index.walkForIndexing(dirEntryPath, origPath, requestPath, config)
		} else if strings.HasSuffix(dirEntry.Name(), config.WatchExtension) {
			index.processUpdate(dirEntryPath, requestPath + dirEntry.Name(), config)
		}
	}
}

// watches a given filepath for an index for changes
func startWatching(filePath string, requestPath string, config IndexSection) WatcherMeta {
	activeWatcher := new(WatcherMeta)
	activeWatcher.watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	activeWatcher.filePath = filePath
	activeWatcher.requestPath = requestPath
	activeWatcher.indexPath = config.IndexPath

	// maybe rework the index so the Watcher is inside the index? idk

	// #TODO this is where my loop isn't stopping
	// #TODO keep rewriting this function (and below) to work differently - per path
	// start a go routine to process events
	go func() {
		idleTimer := time.NewTimer(10 * time.Second)
		queuedEvents := make([]fsnotify.Event, 0)
		for {
			select {
			case ev := <-activeWatcher.watcher.Events:
				queuedEvents = append(queuedEvents, ev)
				idleTimer.Reset(10 * time.Second)
			case err := <-watcher.Errors:
				log.Fatal(err)
			case <-idleTimer.C:
				for _, ev := range queuedEvents {
					if strings.HasSuffix(ev.Name, config.WatchExtension) {
						switch ev.Op {
						case fsnotify.Remove, fsnotify.Rename:
							// delete the filePath
							processDelete(activeWatcher.filePath + ev.Name, activeWatcher.requestPath + ev.Name, config)
						case fsnotify.Create, fsnotify.Write:
							// update the filePath
							processUpdate(activeWatcher.filePath + ev.Name, activeWatcher.requestPath + ev.Name, config)
						default:
							// ignore
						}
					}
				}
				queuedEvents = make([]fsnotify.Event, 0)
				idleTimer.Reset(10 * time.Second)
			}
		}
	}()

	// now actually watch the filePath requested
	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("watching '%s' for changes...", filePath)

	return *watcher
}

// Update the entry in the index to the output from a given file
func (index *GnosisIndex) processUpdate(filePath string, relativePath string, config IndexSection) {
	log.Printf("updated: %s as %s", filePath, relativePath)
	wiki, err := generateWikiFromFile(filePath, config)
	if err != nil {
		log.Print(err)
	} else {
		index, _ := bleve.Open(config.IndexPath)
		defer index.Close()
		index.Index(relativePath, wiki)
	}
}

// Deletes a given path from the wiki entry
func (index *GnosisIndex) processDelete(filePath string, relativePath string, config IndexSection) {
	log.Printf("delete: %s", filePath)
	index, _ := bleve.Open(config.IndexPath)
	defer index.Close()
	index.Index(relativePath, wiki)
	err := index.TrueIndex.Delete(relativePath)
	if err != nil {
		log.Print(err)
	}
}

func (index *GnosisIndex) cleanupMarkdown(input []byte) []byte {
	extensions := 0
	renderer := blackfridaytext.TextRenderer()
	output := blackfriday.Markdown(input, renderer, extensions)
	return output
}

func (index *GnosisIndex) generateWikiFromFile(filePath string, config IndexSection) (*indexedPage, error) {
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filePath)
	//fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if pdata.MatchedTag(config.Restricted) == true {
		return nil, errors.New("Hit a restricted page - " + pdata.Title)
	} else {
		cleanedUpPage := cleanupMarkdown(pdata.Page)
		topics, keywords := pdata.ListMeta()
		rv := indexedPage{
			Name:     pdata.Title,
			Body:     string(cleanedUpPage),
			Filepath: filePath,
			Topics:   strings.Join(topics, " "),
			Keywords: strings.Join(keywords, " "),
			Modified: pdata.FileStats.ModTime(),
		}
		return &rv, nil
	}
}

