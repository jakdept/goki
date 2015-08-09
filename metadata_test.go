// contains the tests for the metadata tests

package gnosis

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	"io/ioutil"
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
	filepath := path.Join(os.TempDir(), "testfile")
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

func TestWriteFileForTest(t *testing.T) {
	stringToWrite := "teststringandstuff"
	filepath := writeFileForTest(t, stringToWrite)
	fileContents, err := ioutil.ReadFile(filepath)
	assert.NoError(t, err, "problem reading file - %v", err)
	assert.Equal(t, stringToWrite, string(fileContents), "file contents did not match")
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
		{1, "", "", "\nTest Page\n",},
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
		assert.Equal(t, testSet.expected, pdata.isOneLineTitle([]byte(testSet.input)),
			"input was [%q]\nwrong amount of characters to discard", testSet.input)
		assert.Equal(t, testSet.expectedTitle, pdata.Title,
			"input was [%q]\ntitle not detected", testSet.input)
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
		assert.Equal(t, testSet.expected, pdata.isTwoLineTitle([]byte(testSet.input)),
			"input was [%q]\nwrong amount of characters to discard", testSet.input)
		assert.Equal(t, testSet.expectedTitle, pdata.Title,
			"input was [%q]\ntitle not detected", testSet.input)
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
		assert.Equal(t, testSet.expected, result,
			"nextLine at wrong pos - expected [%d] but found [%d], input was [%q]",
			testSet.expected, result, testSet.input)
	}
}

func TestCheckMatch(t *testing.T) {
	var checkMatchTests = []struct{
		expected bool
		expectedMatch string
		input string
	}{
		{true, "a", "topic = a",},
		{true, "b", "topic= b",},
		{true, "c", "topic=c",},
		{true, "c", "topic=c",},
		{true, "d-e-f", "topic=d-e-f",},
		{true, "g-h", "topic=g  h",},
		{true, "i", "topic:i",},
		{true, "j", "topic: j",},
		{true, "k", "topic :k",},
		{true, "l-m-no", "topic : l m   no",},
	}

	for _, testSet := range checkMatchTests {
		pdata := new(PageMetadata)
		pdata.checkMatch([]byte(testSet.input), []byte("topic"), &pdata.Topics)
		if testSet.expected {
			assert.True(t, pdata.Topics[testSet.expectedMatch],
				"[%q] - should have seen topic [%q]", testSet.input, testSet.expectedMatch)
		} else {
			assert.False(t, pdata.Topics[testSet.expectedMatch],
				"[%q] - should not have seen topic [%q]", testSet.input, testSet.expectedMatch)
		}
	}
}

func TestProcessMetadata(t *testing.T) {
	var checkProcessMetadata = []struct{
		metaType string
		expected string
		input string
	}{
		{"topic", "a", "topic = a",},
		{"topic", "b", "topic= b",},
		{"topic", "c", "topic=c",},
		{"topic", "c", "topic=c",},
		{"topic", "d-e-f", "topic=d-e-f",},
		{"topic", "g-h", "topic=g  h",},
		{"topic", "i", "topic:i",},
		{"topic", "j", "topic: j",},
		{"topic", "k", "topic :k",},
		{"topic", "l-m-no", "topic : l m   no",},
		{"keyword", "pqrstuv", "keyword : pqrstuv",},
		{"keyword", "lock-box", "keyword:lock             box",},
		{"author", "bob-dole", "author : Bob Dole",},
	}

	for _, testSet := range checkProcessMetadata {
		pdata := new(PageMetadata)
		pdata.processMetadata([]byte(testSet.input))
		switch testSet.metaType {
			case "topic": assert.True(t, pdata.Topics[testSet.expected],
				"[%q] - should have seen topic [%q]", testSet.input, testSet.expected)
			case "keyword": assert.True(t, pdata.Keywords[testSet.expected],
				"[%q] - should have seen topic [%q]", testSet.input, testSet.expected)
			case "author": assert.True(t, pdata.Authors[testSet.expected],
				"[%q] - should have seen topic [%q]", testSet.input, testSet.expected)
		}
	}
}

func TestLoadPage(t *testing.T) {

	var loadPageTests = []struct{
		input string
		expectedOutput string
		expectedTitle string
		expectedKeywords []string
		expectedTopics []string
		expectedAuthors []string
		expectedError string
	}{
		{
			"Test Page\n=========\nsome test content\nand some more",
			"some test content\nand some more",
			"Test Page", []string{}, []string{}, []string{}, "",
		},
		{
			"keyword : junk\nsome other Page\n=========\nsome test content\nthere should be keywords",
			"some test content\nthere should be keywords",
			"some other Page", []string{"junk"}, []string{}, []string{}, "",
		},
		{
			"keyword : junk\nkeyword = other junk\nsome other Page\n=========\nsome test content\nthere should be keywords",
			"some test content\nthere should be keywords",
			"some other Page", []string{"junk","other-junk"}, []string{}, []string{}, "",
		},
		{
			"keyword : junk\nkeyword = other junk\n#some other Page\nsome test content\nthere should be keywords",
			"some test content\nthere should be keywords",
			"some other Page", []string{"junk","other-junk"}, []string{}, []string{}, "",
		},
		{
			"keyword : junk\nkeyword = other junk\ntopic : very important\ncategory=internal\nsome other Page\n=========\nsome test content\nthere should be keywords",
			"some test content\nthere should be keywords",
			"some other Page", []string{"junk", "other-junk"}, []string{"very-important", "internal"}, []string{}, "",
		},
		{
			"========", "", "", []string{}, []string{}, []string{}, "need to hit a title in the file",
		},
		{
			"junk page", "", "", []string{}, []string{}, []string{}, "need to hit a title in the file",
		},
		{
			"junk page\nthat\nhas\nno\ntitle", "", "", []string{}, []string{}, []string{}, "need to hit a title in the file",
		},
		{
			"just a title\n=========", "", "", []string{}, []string{}, []string{}, "need to hit a title in the file",
		},
	}

	var filepath string
	defer os.Remove(filepath)

	for _, testSet := range loadPageTests {
		filepath := writeFileForTest(t, testSet.input)
		pdata := new(PageMetadata)
		err := pdata.LoadPage(filepath)

		if testSet.expectedError != "" {
			assert.Error(t, err, "[%q] should have kicked error - %q")
		} else {	
			assert.NoError(t, err, "[%q] should not have kicked an error, kicked - %q")
		}

		assert.Equal(t, testSet.expectedOutput, string(pdata.Page), "input [%q] page did not match", testSet.input)
		assert.Equal(t, testSet.expectedTitle, string(pdata.Title), "input [%q] title did not match", testSet.input)
		assert.Equal(t, len(testSet.expectedKeywords), len(pdata.Keywords), "input [%q] keyword count did not match", testSet.input)
		assert.Equal(t, len(testSet.expectedTopics), len(pdata.Topics), "input [%q] keyword count did not match", testSet.input)
		assert.Equal(t, len(testSet.expectedAuthors), len(pdata.Authors), "input [%q] keyword count did not match", testSet.input)

		for _, keyword := range testSet.expectedKeywords {
			assert.True(t, pdata.Keywords[keyword], "input [%q] failed to find keyword [%q]", testSet.input, keyword)
		}
		for _, topic := range testSet.expectedTopics {
			assert.True(t, pdata.Topics[topic], "input [%q] failed to find topic [%q]", testSet.input, topic)
		}
		for _, author := range testSet.expectedAuthors {
			assert.True(t, pdata.Authors[author], "input [%q] failed to find author [%q]", testSet.input, author)
		}
	}
}

func TestMatchedTag(t *testing.T) {
	var matchedTagTests = []struct{
		inputTopics []string
		expectedTopics []string
		falseTopics []string
	}{
		{
			[]string{"topic : dog", "topic : baNana", "topic : aPPle", "topic : cat",},
			[]string{"apple", "banana", "cat", "dog",},
			[]string{},
		},
		{
			[]string{"topic : tree frog", "topic : eagle", "topic : goat", "topic : hog",},
			[]string{"eagle", "tree-frog", "goat", "hog",},
			[]string{"apple", "banana", "cat", "dog",},
		},
		{
			[]string{"topic : frog", "topic : eagle", "topic : goat", "topic : hog", "topic : iguana",},
			[]string{},
			[]string{"jester-and-joker", "kangaroo", "llama",},
		},
	}
	for _, testSet := range matchedTagTests {
		pdata := new(PageMetadata)
		pdata.Topics = map[string]bool{}
		for _, topic := range testSet.inputTopics {
			pdata.processMetadata([]byte(topic))
			// pdata.Topics[topic] = true
		}

		for _, singleTopic := range testSet.expectedTopics {
			assert.True(t, pdata.MatchedTopic([]string{singleTopic}),
				"should have found [%q] in [%q]", singleTopic, testSet.inputTopics)
		}
		for _, singleTopic := range testSet.falseTopics {
			assert.False(t, pdata.MatchedTopic([]string{singleTopic}),
				"should not have found [%q] in [%q]", singleTopic, testSet.inputTopics)
		}
	}

}

func TestOldMatchedTag(t *testing.T) {
	// test a page with two topics
	filepath := writeFileForTest(t, "topic : junk\ncategory = other junk\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filepath)
	assert.NoError(t, err)
	assert.True(t, pdata.MatchedTopic([]string{"junk", "not-gonna-hit"}),
		"i couldn't find the expected tag")
	assert.True(t, pdata.MatchedTopic([]string{"other-junk", "not-gonna-hit"}),
		"i couldn't find the expected keyword")
	assert.False(t, pdata.MatchedTopic([]string{"not-added"}),
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

/*
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
*/
