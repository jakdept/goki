// contains the tests for the metadata tests

package gnosis

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func stringKeyExistsInMap(itemMap map[string]bool, key string) bool {
	defer func() bool {
		r := recover()
		return r == nil
	}()
	junk := itemMap[key]
	return junk == junk
}

func TestLineIsTitle(t *testing.T) {
	pdata := new(PageMetadata)

	// test the most normal topic line I'd expect
	titleLine := []byte("=======")
	assert.True(t, pdata.lineIsTitle(titleLine), "the most basic topic line just failed")

	titleLine = []byte("=")
	assert.True(t, pdata.lineIsTitle(titleLine), "one = should be enough")

	titleLine = []byte("   ======")
	assert.True(t, pdata.lineIsTitle(titleLine), "any spaces before the heading portion should not cause failure")

	titleLine = []byte("\t\t\t======")
	assert.True(t, pdata.lineIsTitle(titleLine), "tabs before the heading portion should not cause failure")

	titleLine = []byte("=======     ")
	assert.True(t, pdata.lineIsTitle(titleLine), "spaces after the heading portion should not cause failure")

	titleLine = []byte("=======\t\t\t")
	assert.True(t, pdata.lineIsTitle(titleLine), "tabs after the heading portion should not cause failure")

	titleLine = []byte("=======\n")
	assert.True(t, pdata.lineIsTitle(titleLine), "a newline after the heading portion should not cause failure")

	titleLine = []byte("===== ===")
	assert.False(t, pdata.lineIsTitle(titleLine), "the underlining has to be continous - no spaces - so this should have failed")

	titleLine = []byte("====	=====")
	assert.False(t, pdata.lineIsTitle(titleLine), "the underlining has to be continous - no tabs - so this should have failed")
}

func TestCheckMatch(t *testing.T) {
	pdata := new(PageMetadata)

	metadataLine := []byte("topic = a")
	metadataMatch := []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["a"], "should have been able to add element a to the thingy")

	metadataLine = []byte("topic= b")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["b"], "should have been able to add element b to the thingy")

	metadataLine = []byte("topic=c")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["c"], "should have been able to add element c to the thingy")

	metadataLine = []byte("topic=d e f")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["d-e-f"], "should have been able to add element d e f to the thingy")

	metadataLine = []byte("topic=g  h")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["g-h"], "should have been able to add element g-h to the thingy")

	metadataLine = []byte("topic:i")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["i"], "should have been able to add element i to the thingy")
}

func TestProcessMetadata(t *testing.T) {
	pdata := new(PageMetadata)

	metaDataLine := []byte("topic = a")
	pdata.processMetadata(metaDataLine)
	assert.True(t, pdata.Topics["a"], "topic a should have been added")

	metaDataLine = []byte("keyword=b")
	pdata.processMetadata(metaDataLine)
	assert.True(t, pdata.Keywords["b"], "keyword b should have been added")
}

func TestLoadPage(t *testing.T) {
	// lets test this without any metadata
	simplePage := "Test Page\n=========\nsome test content"

	filepath := path.Join(os.TempDir(), "simplePage.md")
	f, err := os.Create(filepath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath)
	if _, err := f.WriteString(simplePage); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	pdata := new(PageMetadata)
	pdata.LoadPage(filepath)
	assert.Equal(t, pdata.Page, []byte(simplePage), "I should be able to load a page that has no metadata")

	// test a page with a keyword
	keywordPage := "keyword : junk\nsome other Page\n=========\nsome test content\nthere should be keywords"

	filepath = path.Join(os.TempDir(), "simpleKeywordPage.md")
	f, err = os.Create(filepath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath)
	if _, err := f.WriteString(keywordPage); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	// reinitalize the PageMetadata
	pdata = new(PageMetadata)
	pdata.LoadPage(filepath)
	assert.Equal(t, pdata.Page, []byte(keywordPage), "I should be able to load a page that has no metadata")

}
