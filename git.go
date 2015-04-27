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
	"gopkg.in/fsnotify.v1"

"git.drogon.net/wiringPi"
)

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

