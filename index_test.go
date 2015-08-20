
package gnosis

import (
	"github.com/stretchr/testify/assert"
	"os"
	// "bytes"
	"path"
	"path/filepath"
	"testing"
	"log"
	// "time"
	"io/ioutil"
	"strings"
)

func TestGetURIPath(t *testing.T) {
	var tests = []struct{
		input string
		trim string
		add string
		output string
	}{
		{"/wiki.md", "/", "/var/www/", "/var/www/wiki.md"},
		{"/wiki/page.md", "/wiki/", "/var/www/", "/var/www/page.md"},
		{"/wiki/page.md", "/wiki/", "/", "/page.md"},
		{"abcdef", "abc", "xyz", "xyzdef"},
	}

	for _, testSet := range tests {
		assert.Equal(t, testSet.output, getURIPath(testSet.input, testSet.trim, testSet.add),
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

	inputFiles, err := filepath.Glob("./testfiles/*.md")
	if err != nil {
		t.Errorf("Failed to find all markdown files - %s", err)
		return
	}

	for _, input = range inputFiles {
		output := strings.TrimSuffix(input, ".md") + ".cleanupMarkdown"

		rawData, err := ioutil.ReadFile(input)
		if err != nil {
			t.Errorf("Failed to open input file [%q] - %s", input, err)
			return
		}

		expectedData, err := ioutil.ReadFile(output)
		if err != nil {
			t.Errorf("Failed to open input file [%q] - %s", output, err)
			return
		}

		cleanData := cleanupMarkdown(rawData)
		if cleanData != string(expectedData) {
			splitPath:= strings.Split(output, string(os.PathSeparator))
			filepath := path.Join(os.TempDir(), splitPath[len(splitPath)-1])
			f, err := os.Create(filepath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.WriteString(cleanData); err != nil {
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			t.Errorf("Input [%q] gave incorrect output against [%q] - wrote to [%q]",
				input, output, filepath)
		}
	}

}