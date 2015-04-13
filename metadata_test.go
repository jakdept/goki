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

func writeFileForTest(t *testing.T, s string) string {
	filepath := path.Join(os.TempDir(), "testPage.md")
	f, err := os.Create(filepath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(s); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return filepath
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

	pdata.processMetadata([]byte("not really a topic"))
	assert.Equal(t, 0, len(pdata.Topics), "there shouldn't be a topic in there")

	pdata.processMetadata([]byte("topic = a"))
	assert.True(t, pdata.Topics["a"], "topic a should have been added")

	pdata.processMetadata([]byte("keyword=b"))
	assert.True(t, pdata.Keywords["b"], "keyword b should have been added")
}

func TestLoadPage(t *testing.T) {
	// lets test this without any metadata
	filepath := writeFileForTest(t, "Test Page\n=========\nsome test content\nand some more")
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filepath)
	defer os.Remove(filepath)

	assert.NoError(t, err)
	assert.Equal(t, string(pdata.Page),
		"Test Page\n=========\nsome test content\nand some more",
		"I should be able to load a page that has no metadata")

	// test a page with a keyword
	filepath = writeFileForTest(t, "keyword : junk\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	assert.NoError(t, err)
	assert.Equal(t, string(pdata.Page),
		"some other Page\n=========\nsome test content\nthere should be keywords",
		"I should be able to load a page that has a keyword")
	assert.True(t, pdata.Keywords["junk"],
		"i couldn't find the expected keyword")

	// test a page with two keywords
	filepath = writeFileForTest(t, "keyword : junk\nkeyword = other junk\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	assert.NoError(t, err)
	assert.Equal(t, string(pdata.Page),
		"some other Page\n=========\nsome test content\nthere should be keywords",
		"I should be able to load a page that has a keyword")
	assert.True(t, pdata.Keywords["junk"],
		"i couldn't find the expected keyword")
	assert.True(t, pdata.Keywords["other-junk"],
		"i couldn't find the expected keyword")
	assert.False(t, pdata.Keywords["not-added"],
		"i couldn't find the expected keyword")

	filepath = writeFileForTest(t, "keyword : junk\nkeyword = other junk\ntopic : very important\ncategory=internal\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	assert.NoError(t, err)
	assert.Equal(t, string(pdata.Page),
		"some other Page\n=========\nsome test content\nthere should be keywords",
		"I should be able to load a page that has a keyword")
	assert.True(t, pdata.Keywords["junk"],
		"i couldn't find the expected keyword")
	assert.True(t, pdata.Keywords["other-junk"],
		"i couldn't find the expected keyword")
	assert.True(t, pdata.Topics["very-important"],
		"i couldn't find the expected keyword")
	assert.True(t, pdata.Topics["internal"],
		"i couldn't find the expected keyword")
	assert.False(t, pdata.Keywords["not-added"],
		"i couldn't find the expected keyword")
	assert.False(t, pdata.Topics["not-added"],
		"i couldn't find the expected keyword")

	filepath = writeFileForTest(t, "========")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	assert.EqualError(t, err, "I only read in... one line?")
	assert.Empty(t, string(pdata.Page),
		"how did this page gontent get here?")
	assert.False(t, pdata.Keywords["not-added"],
		"i couldn't find the expected keyword")
	assert.False(t, pdata.Topics["not-added"],
		"i couldn't find the expected keyword")

	filepath = writeFileForTest(t, "junk page")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	assert.EqualError(t, err, "I only read in... one line?")
	assert.Empty(t, string(pdata.Page),
		"how did this page gontent get here?")
	assert.False(t, pdata.Keywords["not-added"],
		"i found something I should not have")
	assert.False(t, pdata.Topics["not-added"],
		"i found something I should not have")

	filepath = writeFileForTest(t, "junk page\nthat\nhas\nno\ntitle")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	assert.EqualError(t, err, "I only read in... one line?")
	assert.Empty(t, string(pdata.Page),
		"how did this page gontent get here?")
	assert.False(t, pdata.Keywords["not-added"],
		"i found something I should not have")
	assert.False(t, pdata.Topics["not-added"],
		"i found something I should not have")
}
