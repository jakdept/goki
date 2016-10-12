package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bleve"
	"search"

	"github.com/stretchr/testify/assert"
)

func TestCreateResponseData(t *testing.T) {
	documentMatch := []search.DocumentMatch{
		search.DocumentMatch{
			Score: .87,
			Fields: map[string]interface{}{
				"title":   "test page",
				"path":    "/test_page.md",
				"body":    "test page body",
				"topic":   "a few topics",
				"keyword": "a few keywords",
				"author":  "a few authors",
			},
		},
		search.DocumentMatch{
			Score: .63,
			Fields: map[string]interface{}{
				"title":   "other page",
				"path":    "/other_page.md",
				"body":    "other page body",
				"topic":   "a few topics",
				"keyword": "a few keywords",
				"author":  "a few authors",
			},
		},
	}

	testData := []struct {
		rawResult  bleve.SearchResult
		pageOffset int
		out        SearchResponse
		expErr     error
	}{
		{
			rawResult: bleve.SearchResult{
				Total:    2,
				MaxScore: .87,
				Took:     time.Microsecond * 5,
				Hits: search.DocumentMatchCollection{
					&documentMatch[0],
					&documentMatch[1],
				},
			},
			pageOffset: 0,
			out: SearchResponse{
				TotalHits:  2,
				PageOffset: 0,
				SearchTime: time.Microsecond * 5,
				Topics:     []string{"a", "few", "topics"},
				Authors:    []string{"a", "few", "authors"},
				Results: []SearchResponseResult{
					{},
				},
			},
			expErr: nil,
		},
	}

	if testData[0].pageOffset != 0 {
		t.Fatal("testData was not propagated")
	}
}

func TestGetURIPath(t *testing.T) {
	var tests = []struct {
		input  string
		trim   string
		add    string
		output string
	}{
		{"/wiki.md", "/", "/var/www/", "/var/www/wiki.md"},
		{"/wiki/page.md", "/wiki/", "/var/www/", "/var/www/page.md"},
		{"/wiki/page.md", "/wiki/", "/", "/page.md"},
		{"abc/def", "abc/", "xyz/", "xyz/def"},
	}
	i := &indexObject{}

	for _, testSet := range tests {
		i.config.IndexPath = testSet.add
		assert.Equal(t, testSet.output, i.getURI(testSet.input, testSet.trim, testSet.add),
			"[%q] trimmed [%q] added [%q] but got the wrong result",
			testSet.input, testSet.trim, testSet.add)
	}
}

func TestCleanupMarkdownFiles(t *testing.T) {
	var input string
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("\npanic while processing [%#v]\n", input)
		}
	}()
	i := &indexObject{}

	inputFiles, err := filepath.Glob("./testfiles/*.md")
	if err != nil {
		t.Errorf("Failed to find all markdown files - %s", err)
		return
	}

	for _, input = range inputFiles {
		output := strings.TrimSuffix(input, ".md") + ".cleanupMarkdown"

		rawData, err := ioutil.ReadFile(input)
		if err != nil {
			t.Errorf("failed to open input file [%q] - %s", input, err)
			return
		}

		expectedData, err := ioutil.ReadFile(output)
		if err != nil {
			t.Errorf("failed to open input file [%q] - %s", output, err)
			return
		}

		cleanData := i.cleanupMarkdown(rawData)
		if cleanData != string(expectedData) {
			f, err := ioutil.TempFile("", "cleanupMarkdown.")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write(rawData); err != nil {
				t.Fatal(err)
			}
			t.Errorf("Input [%q] gave incorrect output against [%q] - wrote to [%q]",
				input, output, f.Name())
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

/*
func TestGenerateWikiFromFile(t *testing.T) {
	var input string
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("\npanic while processing [%#v]\n", input)
		}
	}()

	inputFiles, err := filepath.Glob("./testfiles/*.md")
	if err != nil {
		t.Errorf("Failed to find all markdown files - %s", err)
		return
	}

	for _, input = range inputFiles {
		output := strings.TrimSuffix(input, ".md") + ".generateWikiFromFile"

		expectedRaw, err := ioutil.ReadFile(output)
		if err != nil {
			t.Errorf("Failed to open input file [%q] - %s", output, err)
			return
		}

		var expectedValue indexedPage
		err = json.Unmarshal(expectedRaw, &expectedValue)
		if err != nil {
			t.Errorf("Failed to unmarshal the expected values - %s", err)
		}

		expectedRaw, err = json.Marshal(expectedValue)
		if err != nil {
			t.Errorf("Failed to marshal the expected values - %s", err)
		}

		actualValue, err := generateWikiFromFile(input, "teststring", []string{})
		actualRaw, err := json.Marshal(actualValue)
		if err != nil {
			t.Errorf("Failed to marshal the actual values - %s", err)
		}

		if bytes.Compare(actualRaw, expectedRaw) != 0 {
			f, err := ioutil.TempFile("", "generateWikiFromFile.")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write(actualRaw); err != nil {
				t.Fatal(err)
			}
			t.Errorf("Input [%q] gave incorrect output against [%q]\nwrote to [%q]",
				input, output, f.Name())
			t.Errorf("\ninput is  [%s]\noutput is [%s]", actualRaw, expectedRaw)
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}
*/
