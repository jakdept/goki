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

func TestIsTitle(t *testing.T) {

	var isTitleTests = []struct{
		expected int
		expectedTitle string
		expectedTopic string
		input string
	}{
		{12, "title", "", "stuff\n#title",},
		{9, "title", "", "title\n===",},
		{7, "title", "", "#title\n\n",},
		{8, "title", "", "\n#title\n\n",},
		{1, "", "", "\ntitle\n====\n\n",},
		{23, "title", "pageTopic", "topic:pageTopic\n#title\n\n",},
		{22, "title", "pageTopic", "topic:pageTopic\n#title",},
		{16, "", "pageTopic", "topic:pageTopic\ntitle is too late\n====",},
	}

	for _, testSet := range isTitleTests {
		pdata := new(PageMetadata)
		assert.Equal(t, testSet.expected, pdata.isTitle([]byte(testSet.input)), "[%q] - wrong amount of characters discarded", testSet.input)
		assert.Equal(t, testSet.expectedTitle, pdata.Title, "[%q] - title not detected", testSet.input)
		assert.True(t, stringKeyExistsInMap(pdata.Topics, testSet.expectedTopic), "[%q] - metadata key not found", testSet.input)
	}
}

func TestIsOneLineTitle(t *testing.T) {
	var isTitleTests = []struct{
		expected int
		expectedTitle string
		input string
	}{
		{7, "title", "#title\n",},
		{6, "title", "#title",},
		{0, "", "\n#title\n\n",},
		{0, "", "title#\n",},
		{9, "yuup", "# yuup #\nother junk that should not matter",},
		{16, "space stays", "# space stays #\nother junk that should not matter",},
	}

	for _, testSet := range isTitleTests {
		pdata := new(PageMetadata)
		assert.Equal(t, testSet.expected, pdata.isOneLineTitle([]byte(testSet.input)), "input was [%q]\nwrong amount of characters to discard", testSet.input)
		assert.Equal(t, testSet.expectedTitle, pdata.Title, "input was [%q]\ntitle not detected", testSet.input)
	}
}

func TestIsTwoLineTitle(t *testing.T) {
	var isTitleTests = []struct{
		expected int
		expectedTitle string
		input string
	}{
		{0, "", "#title\n",},
		{0, "", "#title",},
		{0, "", "\ntitle\n====\n\n",},
		{11, "title", "title\n====\n\n",},
		{10, "title", "title\n====",},
		{8, "title", "title\n==",},
		{11, "title", "title\n====\nother stuff that should not matter",},
		{22, "three word title", "three word title\n=====",},
		{27, "space before", "\t    \t   space before\n=====",},
		{26, "space after", "space after\t    \t   \n=====",},
	}

	for _, testSet := range isTitleTests {
		pdata := new(PageMetadata)
		assert.Equal(t, testSet.expected, pdata.isTwoLineTitle([]byte(testSet.input)), "input was [%q]\nwrong amount of characters to discard", testSet.input)
		assert.Equal(t, testSet.expectedTitle, pdata.Title, "input was [%q]\ntitle not detected", testSet.input)
	}
}


func TestFindNextLine(t *testing.T) {
	var isTitleTests = []struct{
		expected int
		input string
	}{
		{0, "\n#title\n",},
		{5, "title\n===\n",},
		{-1, "title===",},
		{8, "title===\n",},
	}
	for _, testSet := range isTitleTests {
		pdata := new(PageMetadata)
		result := pdata.findNextLine([]byte(testSet.input))
		assert.Equal(t, testSet.expected, result, "nextLine at wrong pos - expected [%d] but found [%d], input was [%q]", testSet.expected, result, testSet.input)
	}
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
	assert.EqualError(t, err, "never hit a title")
	assert.Empty(t, string(pdata.Page),
		"how did this page gontent get here?")
	assert.False(t, pdata.Keywords["not-added"],
		"i found something I should not have")
	assert.False(t, pdata.Topics["not-added"],
		"i found something I should not have")

	// test for failure with just a file
	filepath = writeFileForTest(t, "just a title\n=========")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	assert.EqualError(t, err, "Is this page just a title?")
}

func TestMatchedTag(t *testing.T) {
	// test a page with two topics
	filepath := writeFileForTest(t, "topic : junk\ncategory = other junk\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filepath)
	assert.NoError(t, err)
	assert.True(t, pdata.MatchedTag([]string{"junk", "not-gonna-hit"}),
		"i couldn't find the expected tag")
	assert.True(t, pdata.MatchedTag([]string{"other-junk", "not-gonna-hit"}),
		"i couldn't find the expected keyword")
	assert.False(t, pdata.MatchedTag([]string{"not-added"}),
		"i couldn't find the expected keyword")
}

func TestListMeta(t *testing.T) {
	// test a page with two keywords and a topic
	filepath := writeFileForTest(t, "keyword : junk\ncategory = other junk\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filepath)
	allTopics, allKeywords, _ := pdata.ListMeta()
	assert.NoError(t, err)
	assert.Equal(t, allTopics, []string{"other-junk"}, "I didn't get the right topic list")
	assert.Equal(t, allKeywords, []string{"junk"}, "I didn't get the right keyword list")

	// test a page with nothing
	filepath = writeFileForTest(t, "some other Page\n=========\nsome test content\nthere should be keywords")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	allTopics, allKeywords, _ = pdata.ListMeta()
	assert.NoError(t, err)
	assert.Equal(t, []string(nil), allTopics, "I didn't get the empty topic list")
	assert.Equal(t, []string(nil), allKeywords, "I didn't get the empty keyword list")
}

func TestPrintTopics(t *testing.T) {
	// test a page with one topics
	filepath := writeFileForTest(t, "keyword : junk\ncategory : other junk\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filepath)
	printedTopics := pdata.PrintTopics("/category/")
	assert.NoError(t, err)
	assert.True(t, pdata.Topics["other-junk"], "The topic isn't directly assigned")
	assert.Equal(t, "<div class='tag'>/category/other-junk</div>", printedTopics, "I didn't get the right topic list")

	// test a page with nothing
	filepath = writeFileForTest(t, "some other Page\n=========\nsome test content\nthere should be keywords")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	printedTopics = pdata.PrintTopics("/")
	assert.NoError(t, err)
	assert.Equal(t, "", printedTopics, "I didn't get the empty topic list")
}

func TestPrintKeywords(t *testing.T) {
	// test a page with two topics
	filepath := writeFileForTest(t, "keyword : junk\ncategory = other junk\ntopic:really junk\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filepath)
	printedKeywords := pdata.PrintKeywords()
	assert.NoError(t, err)
	assert.Equal(t, "<meta name='keywords' content='junk'>", printedKeywords, "I didn't get the right topic list")

	// test a page with nothing
	filepath = writeFileForTest(t, "some other Page\n=========\nsome test content\nthere should be keywords")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	printedKeywords = pdata.PrintKeywords()
	assert.NoError(t, err)
	assert.Equal(t, "", printedKeywords, "I didn't get the empty topic list")
}
