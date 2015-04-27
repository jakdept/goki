package gnosis

import (
	"log"
	"strings"
	"io/ioutil"
	"os"
	"time"
	"crypto/md5"
	"fmt"

"github.com/mschoch/blackfriday-text"
	"github.com/blevesearch/bleve"
	"github.com/libgit2/git2go"
	"gopkg.in/fsnotify.v1"
)

func openIndex(config IndexSection) bleve.Index {
	index, err := bleve.Open(filepath.Clean(config.IndexPath))
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Creating new index...")
		// create a mapping
		indexMapping := buildIndexMapping(config)
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
	indexMapping.AddDocumentMapping("wiki", wikiMapping)
	indexMapping.DefaultAnalyzer = config.IndexType

	return indexMapping
}

func processUpdate(index bleve.Index, repo *git.Repository, path string) {
	log.Printf("updated: %s", path)
	rp := relativePath(path)
	wiki, err := NewWikiFromFile(path)
	if err != nil {
		log.Print(err)
	} else {
		doGitStuff(repo, rp, wiki)
		index.Index(rp, wiki)
	}
}

func processDelete(index bleve.Index, repo *git.Repository, path string) {
	log.Printf("delete: %s", path)
	rp := relativePath(path)
	err := index.Delete(rp)
	if err != nil {
		log.Print(err)
	}
}

func relativePath(path string) string {
	if strings.HasPrefix(path, *dir) {
		path = path[len(*dir)+1:]
	}
	return path
}

func walkForIndexing(path string, index bleve.Index, repo *git.Repository) {

	dirEntries, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, dirEntry := range dirEntries {
		dirEntryPath := path + string(os.PathSeparator) + dirEntry.Name()
		if dirEntry.IsDir() {
			walkForIndexing(dirEntryPath, index, repo)
		} else if pathMatch(dirEntry.Name()) {
			processUpdate(index, repo, dirEntryPath)
		}
	}
}


func NewWikiFromFile(path string) (*WikiPage, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cleanedUpBytes := cleanupMarkdown(fileBytes)

	name := path
	lastSlash := strings.LastIndex(path, string(os.PathSeparator))
	if lastSlash > 0 {
		name = name[lastSlash+1:]
	}
	if strings.HasSuffix(name, ".md") {
		name = name[0 : len(name)-len(".md")]
	}
	rv := WikiPage{
		Name: name,
		Body: string(cleanedUpBytes),
	}
	return &rv, nil
}

func cleanupMarkdown(input []byte) []byte {
	extensions := 0
	renderer := blackfridaytext.TextRenderer()
	output := blackfriday.Markdown(input, renderer, extensions)
	return output
}

func startWatching(path string, index bleve.Index, repo *git.Repository) *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

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
					if pathMatch(ev.Name) {
						switch ev.Op {
						case fsnotify.Remove, fsnotify.Rename:
							// delete the path
							processDelete(index, repo, ev.Name)
						case fsnotify.Create, fsnotify.Write:
							// update the path
							processUpdate(index, repo, ev.Name)
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

	// now actually watch the path requested
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("watching '%s' for changes...", path)

	return watcher
}

func openGitRepo(path string) *git.Repository {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Fatal(err)
	}

	return repo
}

func doGitStuff(repo *git.Repository, path string, wiki *WikiPage) {

	// lookup head
	head, err := repo.Head()
	if err != nil {
		log.Print(err)
	} else {
		// lookup commit object
		headOid := head.Target()
		commit, err := repo.LookupCommit(headOid)
		if err != nil {
			log.Print(err)
		}

		// start diffing backwards
		diffCommit, err := recursiveDiffLookingForFile(repo, commit, path)
		if err != nil {
			log.Print(err)
		} else if diffCommit != nil {
			author := diffCommit.Author()
			wiki.ModifiedByName = author.Name
			wiki.ModifiedByEmail = author.Email
			wiki.Modified = author.When
			if wiki.ModifiedByEmail != "" {
				wiki.ModifiedByGravatar = gravatarHashFromEmail(wiki.ModifiedByEmail)
				log.Printf("gravatar hash is: %s", wiki.ModifiedByGravatar)
			}
		} else {
			log.Printf("unable to find commit where file changed")
		}
	}
}

func recursiveDiffLookingForFile(repo *git.Repository, commit *git.Commit, path string) (*git.Commit, error) {
	log.Printf("checking commit %s", commit.Id())
	// if there is a parent, diff against it
	// totally not going to think about branches
	if commit.ParentCount() > 0 {
		parent := commit.Parent(0)

		found := false
		dcb := func(dd git.DiffDelta, x float64) (git.DiffForEachHunkCallback, error) {
			if dd.NewFile.Path == path {
				found = true
			} else if dd.OldFile.Path == path {
				found = true
			}
			return nil, nil
		}

		parentTree, err := parent.Tree()
		if err != nil {
			return nil, err
		}
		commitTree, err := commit.Tree()
		if err != nil {
			return nil, err
		}
		diffOptions, err := git.DefaultDiffOptions()
		if err != nil {
			return nil, err
		}
		diff, err := repo.DiffTreeToTree(parentTree, commitTree, &diffOptions)
		if err != nil {
			return nil, err
		} else {
			diff.ForEach(dcb, git.DiffDetailFiles)
			if found {
				return commit, nil
			} else {
				return recursiveDiffLookingForFile(repo, parent, path)
			}
		}
	} else {
		// if there is no parent check to see if this file
		// was in the commit, if so, this is its
		commitTree, err := commit.Tree()
		if err != nil {
			return nil, err
		}
		treeEntry := commitTree.EntryByName(path)
		if treeEntry != nil {
			return commit, nil
		}
		return nil, nil
	}
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
