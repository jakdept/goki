package gnosis

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
	"errors"

	"github.com/JackKnifed/blackfriday"
	"github.com/blevesearch/bleve"
	"github.com/mschoch/blackfriday-text"
	"gopkg.in/fsnotify.v1"
)

type gnosisIndex struct {
	TrueIndex bleve.Index
	Config    IndexSection
	// ##TODO## figure this out
	// I cannot use append and range on an array of pointers
	// and yet I do not think I can modify an array of non-pointers
	// IDK what to do just yet
	openWatchers []fsnotify.Watcher
}

type indexedPage struct {
	Name     string    `json:"name"`
	Filepath string    `json:"path"`
	Body     string    `json:"body"`
	Topics   string    `json:"topic"`
	Keywords string    `json:"keyword"`
	Modified time.Time `json:"modified"`
}

func openIndex(config IndexSection) (*gnosisIndex, error) {
	index := new(gnosisIndex)
	index.Config = config
	newIndex, err := bleve.Open(path.Clean(index.Config.IndexPath))
	if err == nil {
		log.Printf("Opening existing index %s", index.Config.IndexName)
		index.TrueIndex = newIndex
	} else if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Creating new index %s", index.Config.IndexName)
		// create a mapping
		indexMapping := index.buildIndexMapping()
		newIndex, err := bleve.New(path.Clean(index.Config.IndexPath), indexMapping)
		if err != nil {
			log.Fatal(err)
		} else {
			index.TrueIndex = newIndex
		}
	} else {
		log.Printf("Got an error opening an index %v", err)
		return nil, err
	}
	// You only got here if you opened an existing index, or a new index
	for _, dir := range index.Config.WatchDir {
		log.Printf("Watching and walking dir %s index %s", dir, index.Config.IndexName)
		watcher := index.startWatching(dir)
		index.openWatchers = append(index.openWatchers, watcher)
		index.walkForIndexing(dir, dir)
	}
	return index, nil
}

func (index *gnosisIndex) closeIndex() {
	for _, watcher := range index.openWatchers {
		watcher.Close()
	}
}

func (index *gnosisIndex) buildIndexMapping() *bleve.IndexMapping {

	// create a text field type
	enTextFieldMapping := bleve.NewTextFieldMapping()
	enTextFieldMapping.Analyzer = index.Config.IndexType

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
	indexMapping.AddDocumentMapping(index.Config.IndexName, wikiMapping)
	indexMapping.DefaultAnalyzer = index.Config.IndexType

	return indexMapping
}

func (index *gnosisIndex) cleanupMarkdown(input []byte) []byte {
	extensions := 0
	renderer := blackfridaytext.TextRenderer()
	output := blackfriday.Markdown(input, renderer, extensions)
	return output
}

func (index *gnosisIndex) relativePath(filePath string, dir string) string {
	filePath = strings.TrimPrefix(filePath, dir)
	filePath = path.Clean(filePath)
	return filePath
}

func (index *gnosisIndex) generateWikiFromFile(filePath string) (*indexedPage, error) {
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filePath)
	//fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if pdata.MatchedTag(index.Config.Restricted) == true {
		return nil, errors.New("Hit a restricted page - " + pdata.Title);
	} else {
	cleanedUpPage := index.cleanupMarkdown(pdata.Page)
	topics, keywords := pdata.ListMeta()
	rv := indexedPage{
		Name: pdata.Title,
		Body: string(cleanedUpPage),
		Filepath: filePath,
		Topics: strings.Join(topics, " "),
		Keywords: strings.Join(keywords, " "),
		Modified: pdata.FileStats.ModTime(),
	}
	return &rv, nil
	}
}

// Update the entry in the index to the output from a given file
func (index *gnosisIndex) processUpdate(path string, dir string) {
	log.Printf("updated: %s", path)
	rp := index.relativePath(path, dir)
	wiki, err := index.generateWikiFromFile(path)
	if err != nil {
		log.Print(err)
	} else {
		index.TrueIndex.Index(rp, wiki)
	}
}

// Deletes a given path from the wiki entry
func (index *gnosisIndex) processDelete(path string, dir string) {
	log.Printf("delete: %s", path)
	rp := index.relativePath(path, dir)
	err := index.TrueIndex.Delete(rp)
	if err != nil {
		log.Print(err)
	}
}

// walks a given path, and runs processUpdate on each File
func (index *gnosisIndex) walkForIndexing(path string, origPath string) {
	dirEntries, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, dirEntry := range dirEntries {
		dirEntryPath := path + string(os.PathSeparator) + dirEntry.Name()
		if dirEntry.IsDir() {
			index.walkForIndexing(dirEntryPath, origPath)
		} else if strings.HasSuffix(dirEntry.Name(), index.Config.WatchExtension) {
			index.processUpdate(dirEntryPath, origPath)
		}
	}
}

// watches a given filepath for an index for changes
func (index *gnosisIndex) startWatching(filePath string) fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// maybe rework the index so the Watcher is inside the index? idk

	// start a go routine to process events
	go func() {
		idleTimer := time.NewTimer(10 * time.Second)
		queuedEvents := make([]fsnotify.Event, 0)
		for {
			select {
			case ev := <-watcher.Events:
				queuedEvents = append(queuedEvents, ev)
				idleTimer.Reset(10 * time.Second)
			case err := <-watcher.Errors:
				log.Fatal(err)
			case <-idleTimer.C:
				for _, ev := range queuedEvents {
					if strings.HasSuffix(ev.Name, index.Config.WatchExtension) {
						switch ev.Op {
						case fsnotify.Remove, fsnotify.Rename:
							// delete the filePath
							index.processDelete(ev.Name, filePath)
						case fsnotify.Create, fsnotify.Write:
							// update the filePath
							index.processUpdate(ev.Name, filePath)
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
