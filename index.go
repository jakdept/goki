package gnosis

import (
	"log"
	"strings"
	"io/ioutil"
	"os"
	"time"
	//"crypto/md5"
	"path"
	//"fmt"

"github.com/mschoch/blackfriday-text"
	"github.com/blevesearch/bleve"
	"gopkg.in/fsnotify.v1"
)

type gnosisIndex struct {
	Index bleve.Index
	Config IndexSection
	Repo *git.Repository
}

func openIndex(config IndexSection) bleve.Index {
	index, err := bleve.Open(filepath.Clean(config.IndexPath))
	index.Config = config
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Creating new index...")
		// create a mapping
		indexMapping := index.buildIndexMapping(config)
		index, err = bleve.New(filepath.Clean(config.IndexPath), indexMapping)
		if err != nil {
			log.Fatal(err)
		}
	} else if err == nil {
		log.Printf("Opening existing index...")
	} else {
		log.Fatal(err)
	}
	return index
}

func (index *bleve.Index) buildIndexMapping() *bleve.IndexMapping {

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
	indexMapping.AddDocumentMapping("wiki", wikiMapping)
	indexMapping.DefaultAnalyzer = index.Config.IndexType

	return indexMapping
}

func (index *gnosisIndex) processUpdate(path string) {
	log.Printf("updated: %s", path)
	rp := index.relativePath(path)
	wiki, err := index.generateWikiFromFile(path)
	if err != nil {
		log.Print(err)
	} else {
		doGitStuff(index.Repo, rp, wiki)
		index.Index(rp, wiki)
	}
}

func (index *gnosisIndex) processDelete(path string) {
	log.Printf("delete: %s", path)
	rp := index.relativePath(path)
	err := index.Delete(rp)
	if err != nil {
		log.Print(err)
	}
}

func (index gnosisIndex) relativePath(filePath string) string {
	filePath = strings.TrimPrefix(path, config.WatchDir)
	filePath = path.Clean(path)
	return path
}

func (index *gnosisIndex) walkForIndexing(path string) {

	dirEntries, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, dirEntry := range dirEntries {
		dirEntryPath := path + string(os.PathSeparator) + dirEntry.Name()
		if dirEntry.IsDir() {
			index.walkForIndexing(dirEntryPath)
		} else if strings.HasSuffix(dirEntry.Name(), ".md") {
			index.processUpdate(dirEntryPath)
		}
	}
}


func (index *gnosisIndex) generateWikiFromFile(filePath string) (*WikiPage, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// ##TODO## I need to look up the actual index that I'm building and hit all of the fields here
	cleanedUpBytes := index.cleanupMarkdown(fileBytes)
	name := path.Base(filePath)
	name = strings.TrimSuffix(name, ".md")
	rv := WikiPage{
		Name: name,
		Body: string(cleanedUpBytes),
	}
	return &rv, nil
}

func (index *gnosisIndex) cleanupMarkdown(input []byte) []byte {
	extensions := 0
	renderer := blackfridaytext.TextRenderer()
	output := blackfriday.Markdown(input, renderer, extensions)
	return output
}

func (index *gnosisIndex) startWatching(filePath string) *fsnotify.Watcher {
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
					if strings.HasSuffix(ev.Name, ".md") {
						switch ev.Op {
						case fsnotify.Remove, fsnotify.Rename:
							// delete the filePath
							index.processDelete(index.Index, repo, ev.Name)
						case fsnotify.Create, fsnotify.Write:
							// update the filePath
							index.processUpdate(index.Index, repo, ev.Name)
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

	return watcher
}

/*
func gravatarHashFromEmail(email string) string {
	input := strings.ToLower(strings.TrimSpace(email))
	return fmt.Sprintf("%x", md5.Sum([]byte(input)))
}
*/

/*
func (w *WikiPage) Type() string {
	return "wiki"
}
*/
