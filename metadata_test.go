// contains the tests for the metadata tests

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

// func TestIsTitle(t *testing.T) {

// 	var isTitleTests = []struct {
// 		expected      int
// 		expectedTitle string
// 		expectedTopic string
// 		input         string
// 	}{
// 		{12, "title", "", "stuff\n#title"},
// 		{9, "title", "", "title\n==="},
// 		{7, "title", "", "#title\n\n"},
// 		{8, "title", "", "\n#title\n\n"},
// 		{1, "", "", "\ntitle\n====\n\n"},
// 		{23, "title", "pageTopic", "topic:pageTopic\n#title\n\n"},
// 		{22, "title", "pageTopic", "topic:pageTopic\n#title"},
// 		{16, "", "pageTopic", "topic:pageTopic\ntitle is too late\n===="},
// 		{1, "", "", "\nTest Page\n"},
// 	}

// 	for _, testSet := range isTitleTests {
// 		pdata := new(PageMetadata)
// 		assert.Equal(t, testSet.expected, pdata.isTitle([]byte(testSet.input)), "[%q] - wrong amount of characters discarded", testSet.input)
// 		assert.Equal(t, testSet.expectedTitle, pdata.Title, "[%q] - title not detected", testSet.input)
// 		assert.True(t, stringKeyExistsInMap(pdata.Topics, testSet.expectedTopic), "[%q] - metadata key not found", testSet.input)
// 	}
// }

// func TestIsOneLineTitle(t *testing.T) {
// 	var isTitleTests = []struct {
// 		expected      int
// 		expectedTitle string
// 		input         string
// 	}{
// 		{7, "title", "#title\n"},
// 		{6, "title", "#title"},
// 		{0, "", "\n#title\n\n"},
// 		{0, "", "title#\n"},
// 		{9, "yuup", "# yuup #\nother junk that should not matter"},
// 		{16, "space stays", "# space stays #\nother junk that should not matter"},
// 	}

// 	for _, testSet := range isTitleTests {
// 		pdata := new(PageMetadata)
// 		assert.Equal(t, testSet.expected, pdata.isOneLineTitle([]byte(testSet.input)),
// 			"input was [%q]\nwrong amount of characters to discard", testSet.input)
// 		assert.Equal(t, testSet.expectedTitle, pdata.Title,
// 			"input was [%q]\ntitle not detected", testSet.input)
// 	}
// }

// func TestIsTwoLineTitle(t *testing.T) {
// 	var isTitleTests = []struct {
// 		expected      int
// 		expectedTitle string
// 		input         string
// 	}{
// 		{0, "", "#title\n"},
// 		{0, "", "#title"},
// 		{0, "", "\ntitle\n====\n\n"},
// 		{11, "title", "title\n====\n\n"},
// 		{10, "title", "title\n===="},
// 		{8, "title", "title\n=="},
// 		{11, "title", "title\n====\nother stuff that should not matter"},
// 		{22, "three word title", "three word title\n====="},
// 		{27, "space before", "\t    \t   space before\n====="},
// 		{26, "space after", "space after\t    \t   \n====="},
// 	}

// 	for _, testSet := range isTitleTests {
// 		pdata := new(PageMetadata)
// 		assert.Equal(t, testSet.expected, pdata.isTwoLineTitle([]byte(testSet.input)),
// 			"input was [%q]\nwrong amount of characters to discard", testSet.input)
// 		assert.Equal(t, testSet.expectedTitle, pdata.Title,
// 			"input was [%q]\ntitle not detected", testSet.input)
// 	}
// }

func TestFindNextLine(t *testing.T) {
	var isTitleTests = []struct {
		expected int
		input    string
	}{
		{0, "\n#title\n"},
		{5, "title\n===\n"},
		{-1, "title==="},
		{8, "title===\n"},
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
	var checkMatchTests = []struct {
		expected      bool
		expectedMatch string
		input         string
	}{
		{true, "a", "topic = a"},
		{true, "b", "topic= b"},
		{true, "c", "topic=c"},
		{true, "c", "topic=c"},
		{true, "d-e-f", "topic=d-e-f"},
		{true, "g-h", "topic=g  h"},
		{true, "i", "topic:i"},
		{true, "j", "topic: j"},
		{true, "k", "topic :k"},
		{true, "l-m-no", "topic : l m   no"},
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

// func TestProcessMetadata(t *testing.T) {
// 	var checkProcessMetadata = []struct {
// 		metaType string
// 		expected string
// 		input    string
// 	}{
// 		{"topic", "a", "topic = a"},
// 		{"topic", "b", "topic= b"},
// 		{"topic", "c", "topic=c"},
// 		{"topic", "c", "topic=c"},
// 		{"topic", "d-e-f", "topic=d-e-f"},
// 		{"topic", "g-h", "topic=g  h"},
// 		{"topic", "i", "topic:i"},
// 		{"topic", "j", "topic: j"},
// 		{"topic", "k", "topic :k"},
// 		{"topic", "l-m-no", "topic : l m   no"},
// 		{"keyword", "pqrstuv", "keyword : pqrstuv"},
// 		{"keyword", "lock-box", "keyword:lock             box"},
// 		{"author", "bob-dole", "author : Bob Dole"},
// 	}

// 	for _, testSet := range checkProcessMetadata {
// 		pdata := new(PageMetadata)
// 		pdata.processMetadata([]byte(testSet.input))
// 		switch testSet.metaType {
// 		case "topic":
// 			assert.True(t, pdata.Topics[testSet.expected],
// 				"[%q] - should have seen topic [%q]", testSet.input, testSet.expected)
// 		case "keyword":
// 			assert.True(t, pdata.Keywords[testSet.expected],
// 				"[%q] - should have seen topic [%q]", testSet.input, testSet.expected)
// 		case "author":
// 			assert.True(t, pdata.Authors[testSet.expected],
// 				"[%q] - should have seen topic [%q]", testSet.input, testSet.expected)
// 		}
// 	}
// }

func TestLoadPage(t *testing.T) {

	var loadPageTests = []struct {
		input            string
		expectedOutput   string
		expectedTitle    string
		expectedKeywords []string
		expectedTopics   []string
		expectedAuthors  []string
		expectedError    string
	}{
		{
			"\n\nTest Page\n=========\nsome test content\nand some more",
			"\nTest Page\n=========\nsome test content\nand some more",
			"/Users/jack/tmp/testfile", []string{}, []string{}, []string{}, "",
		},
		{
			"keyword: junk\n\nsome other Page\n=========\nsome test content\nthere should be keywords",
			"some other Page\n=========\nsome test content\nthere should be keywords",
			"/Users/jack/tmp/testfile", []string{"junk"}, []string{}, []string{}, "",
		},
		{
			"keyword: junk\nkeyword: other junk\n\nsome other Page\n=========\nsome test content\nthere should be keywords",
			"some other Page\n=========\nsome test content\nthere should be keywords",
			"/Users/jack/tmp/testfile", []string{"junk", "other junk"}, []string{}, []string{}, "",
		},
		{
			"keyword: junk\nkeyword: other junk\n\n#some other Page\nsome test content\nthere should be keywords",
			"#some other Page\nsome test content\nthere should be keywords",
			"/Users/jack/tmp/testfile", []string{"junk", "other junk"}, []string{}, []string{}, "",
		},
		{
			"keyword: junk\nkeyword: other junk\ntopic: very important\ntopic: internal\n\nsome other Page\n=========\nsome test content\nthere should be keywords",
			"some other Page\n=========\nsome test content\nthere should be keywords",
			"/Users/jack/tmp/testfile", []string{"junk", "other junk"}, []string{"very important", "internal"}, []string{}, "",
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

	for testId, testSet := range loadPageTests {
		filepath := writeFileForTest(t, testSet.input)
		pdata := new(PageMetadata)
		err := pdata.LoadPage(filepath)

		if testSet.expectedError != "" {
			assert.Error(t, err, "[%q] should have kicked error - %q")
		} else {
			assert.NoError(t, err, "[%q] should not have kicked an error, kicked - %q")
		}

		assert.Equal(t, testSet.expectedOutput, string(pdata.Page), "test #%d input [%q] page did not match", testId, testSet.input)
		assert.Equal(t, testSet.expectedTitle, string(pdata.Title), "test #%d input [%q] title did not match", testId, testSet.input)
		assert.Equal(t, len(testSet.expectedKeywords), len(pdata.Keywords), "test #%d input [%q] keyword count did not match", testId, testSet.input)
		assert.Equal(t, len(testSet.expectedTopics), len(pdata.Topics), "test #%d input [%q] keyword count did not match", testId, testSet.input)
		assert.Equal(t, len(testSet.expectedAuthors), len(pdata.Authors), "test #%d input [%q] keyword count did not match", testId, testSet.input)

		for _, keyword := range testSet.expectedKeywords {
			assert.True(t, pdata.Keywords[keyword], "test #%d input failed to find keyword [%q]",
				testId, keyword)
		}
		for _, topic := range testSet.expectedTopics {
			assert.True(t, pdata.Topics[topic], "test #%d failed to find topic [%q]",
				testId, topic)
		}
		for _, author := range testSet.expectedAuthors {
			assert.True(t, pdata.Authors[author], "test #%d failed to find author [%q]",
				testId, author)
		}
	}
}

func TestMatchedTag(t *testing.T) {
	var matchedTagTests = []struct {
		inputTopics    map[string]bool
		expectedTopics []string
		falseTopics    []string
	}{
		{
			map[string]bool{"dog": true, "banana": true, "apple": true, "cat": true},
			[]string{"apple", "banana", "cat", "dog"},
			[]string{},
		},
		{
			map[string]bool{"tree frog": true, "eagle": true, "goat": true, "hog": true},
			[]string{"eagle", "tree frog", "goat", "hog"},
			[]string{"apple", "banana", "cat", "dog"},
		},
		{
			map[string]bool{"frog": true, "eagle": true, "goat": true, "hog": true, "iguana": true},
			[]string{},
			[]string{"jester and joker", "kangaroo", "llama"},
		},
	}
	for _, testSet := range matchedTagTests {
		pdata := new(PageMetadata)
		pdata.Topics = testSet.inputTopics

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

func TestListMeta(t *testing.T) {
	// test a page with two keywords and a topic
	filepath := writeFileForTest(t, "keyword: junk\ntopic: other junk\n\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata := new(PageMetadata)
	err := pdata.LoadPage(filepath)
	allTopics, allKeywords, _ := pdata.ListMeta()
	assert.NoError(t, err)
	assert.Equal(t, []string{"other junk"}, allTopics, "I didn't get the right topic list")
	assert.Equal(t, []string{"junk"}, allKeywords, "I didn't get the right keyword list")

	// test a page with nothing
	filepath = writeFileForTest(t, "\n\nsome other Page\n=========\nsome test content\nthere should be keywords")
	pdata = new(PageMetadata)
	err = pdata.LoadPage(filepath)
	allTopics, allKeywords, _ = pdata.ListMeta()
	assert.NoError(t, err)
	assert.Equal(t, []string(nil), allTopics, "I didn't get the empty topic list")
	assert.Equal(t, []string(nil), allKeywords, "I didn't get the empty keyword list")
}

func TestBodyParseMarkdown(t *testing.T) {
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
		output := strings.TrimSuffix(input, ".md") + ".bodyParseMarkdown"

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

		cleanData := bodyParseMarkdown(rawData)
		if bytes.Compare(cleanData, expectedData) != 0 {
			f, err := ioutil.TempFile("", "bodyParseMarkdown.")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write(cleanData); err != nil {
				t.Fatal(err)
			}
			t.Errorf("Input [%q] gave incorrect output against [%q]\nwrote to [%q]",
				input, output, f.Name())
			t.Errorf("actual [%q]\nexpected [%q]", cleanData, expectedData)
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestTocParseMarkdown(t *testing.T) {
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
		output := strings.TrimSuffix(input, ".md") + ".tocParseMarkdown"

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

		cleanData := tocParseMarkdown(rawData)
		if bytes.Compare(cleanData, expectedData) != 0 {
			f, err := ioutil.TempFile("", "tocParseMarkdown.")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write(cleanData); err != nil {
				t.Fatal(err)
			}
			t.Errorf("Input [%q] gave incorrect output against [%q]\nwrote to [%q]",
				input, output, f.Name())
			t.Errorf("actual [%q]\nexpected [%q]", cleanData, expectedData)
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}
