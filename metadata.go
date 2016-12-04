// this file contains a pre-processor to pull some stuff out of the markdown file before parsing it

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"os"
	"sort"

	"github.com/JackKnifed/blackfriday"
	tocRenderer "github.com/JackKnifed/goki/tocRenderer"
)

type PageMetadata struct {
	Keywords  map[string]bool
	Topics    map[string]bool
	Authors   map[string]bool
	Page      []byte
	Title     string
	FileStats os.FileInfo
}

const (
	bodyHtmlFlags = 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES |
		blackfriday.HTML_ALERT_BOXES

	bodyExtensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_AUTO_HEADER_IDS |
		blackfriday.EXTENSION_TITLEBLOCK |
		blackfriday.EXTENSION_ALERT_BOXES

	tocHtmlFlags = 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES |
		blackfriday.HTML_TOC |
		blackfriday.HTML_OMIT_CONTENTS |
		blackfriday.HTML_FLAT_TOC

	tocExtensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_AUTO_HEADER_IDS |
		blackfriday.EXTENSION_TITLEBLOCK
)

// // take a given line, and check it against every possible type of tag
// func (pdata *PageMetadata) processMetadata(line []byte) {
// 	pdata.checkMatch(line, []byte("tag"), &pdata.Topics)
// 	pdata.checkMatch(line, []byte("topic"), &pdata.Topics)
// 	pdata.checkMatch(line, []byte("category"), &pdata.Topics)

// 	pdata.checkMatch(line, []byte("keyword"), &pdata.Keywords)
// 	pdata.checkMatch(line, []byte("meta"), &pdata.Keywords)

// 	pdata.checkMatch(line, []byte("author"), &pdata.Authors)
// 	pdata.checkMatch(line, []byte("maintainer"), &pdata.Authors)
// }

func (pdata *PageMetadata) checkMatch(
	input []byte, looking []byte, tracker *map[string]bool) {
	// trim off any blank spaces at the start of the line
	input = bytes.ToLower(bytes.TrimSpace(input))
	looking = bytes.ToLower(bytes.TrimSpace(looking))

	if !bytes.HasPrefix(input, looking) {
		return
	}

	input = bytes.TrimSpace(bytes.TrimPrefix(input, looking))

	if input[0] != '=' && input[0] != ':' {
		return
	}

	input = bytes.TrimSpace(input[1:])

	if input[0] == '=' || input[0] == ':' {
		return
	}

	input = bytes.Replace(input, []byte("\t"), []byte(" "), -1)

	parts := bytes.Split(input, []byte(" "))
	var cleanParts [][]byte

	for _, piece := range parts {
		if len(piece) > 0 {
			cleanParts = append(cleanParts[:], piece)
		}
	}

	key := bytes.Join(cleanParts, []byte("-"))

	if *tracker != nil {
		(*tracker)[string(key)] = true
	} else {
		*tracker = map[string]bool{string(key): true}
	}
}

func convertArr(in []string) map[string]bool {
	out := make(map[string]bool)
	log.Printf("\n%#v", in)
	for _, each := range in {
		out[each] = true
	}
	return out
}

func (pdata *PageMetadata) LoadPage(pageName string) error {
	// open the file
	f, err := os.Open(pageName)
	if err != nil {
		return fmt.Errorf("problem opening file [%q] - %v", pageName, err)
	}
	defer f.Close()
	pdata.FileStats, err = os.Stat(pageName)

	parsed, err := mail.ReadMessage(f)
	if err != nil {
		return fmt.Errorf("problem parsing page [%q] - %v", pageName, err)
	}

	log.Printf("\n%#v\n", parsed.Header)

	pdata.Topics = convertArr(parsed.Header["Topic"])
	pdata.Authors = convertArr(parsed.Header["Author"])
	pdata.Keywords = convertArr(parsed.Header["Keyword"])
	if len(parsed.Header["title"]) > 0 {
		pdata.Title = parsed.Header["Title"][0]
	} else {
		pdata.Title = pageName
	}

	tempPage, err := ioutil.ReadAll(parsed.Body)
	if err != nil {
		return fmt.Errorf("problem reading page from parsed version - %v", err)
	}

	pdata.Page = tempPage

	return nil
}

// // determines if the next two lines contain a title line
// // if the first line is not a line, treat it as metadata
// // return the amount of characters processed if not a new line
// // if title line, return total length of the input
// func (pdata *PageMetadata) isTitle(input []byte) int {
// 	newline := []byte("\n")
// 	nextLine := bytes.Index(input, newline)
// 	if nextLine == -1 {
// 		return pdata.isOneLineTitle(input)
// 	}
// 	if rv := pdata.isOneLineTitle(input[:nextLine+1]); rv != 0 {
// 		return rv
// 	}
// 	if nextLine >= len(input) {
// 		return 0
// 	}

// 	lineAfter := bytes.Index(input[nextLine+1:], newline)
// 	if lineAfter == -1 {
// 		lineAfter = len(input) - 1
// 	} else {
// 		lineAfter += nextLine + 1
// 	}

// 	if rv := pdata.isTwoLineTitle(input[:lineAfter+1]); rv != 0 {
// 		return rv
// 	}

// 	pdata.processMetadata(input[:nextLine])
// 	if rv := pdata.isOneLineTitle(input[nextLine+1 : lineAfter+1]); rv != 0 {
// 		return rv + nextLine + 1
// 	} else {
// 		return nextLine + 1
// 	}
// }

// // checks to see if the first lines of a []byte contain a markdown title
// // returns the number of characters to lose
// // 0 indicates failure (no characters to lose)
// func (pdata *PageMetadata) isOneLineTitle(input []byte) int {
// 	var singleLine []byte
// 	var endOfLine int

// 	if endOfLine = pdata.findNextLine(input); endOfLine != -1 {
// 		singleLine = bytes.TrimSpace(input[:endOfLine])
// 	} else {
// 		endOfLine = len(input) - 1
// 		singleLine = bytes.TrimSpace(input)
// 	}

// 	if len(singleLine) > 2 && singleLine[0] == '#' && singleLine[1] != '#' {
// 		singleLine = bytes.Trim(singleLine, "#")
// 		pdata.Title = string(bytes.TrimSpace(singleLine))
// 		return endOfLine + 1
// 	}
// 	return 0
// }

// // checks to see if the first two lines of a []byte contain a markdown title
// // returns the number of characters to lose
// // 0 indicates failure (no characters to lose)
// func (pdata *PageMetadata) isTwoLineTitle(input []byte) int {
// 	var firstNewLine, secondNewLine int

// 	if firstNewLine = pdata.findNextLine(input); firstNewLine == -1 {
// 		return 0
// 	}
// 	secondNewLine = pdata.findNextLine(input[firstNewLine+1:])
// 	if secondNewLine == -1 {
// 		secondNewLine = len(input) - 1
// 	} else {
// 		secondNewLine += firstNewLine + 1
// 	}

// 	secondLine := bytes.TrimSpace(input[firstNewLine+1 : secondNewLine+1])
// 	if len(secondLine) >= 2 {
// 		secondLine = bytes.TrimLeft(secondLine, "=")
// 		if len(secondLine) == 0 {
// 			pdata.Title = string(bytes.TrimSpace(input[:firstNewLine]))
// 			return secondNewLine + 1
// 		}
// 	}
// 	return 0
// }

// given input, find where the next line starts
func (pdata *PageMetadata) findNextLine(input []byte) int {
	nextLine := 0
	for nextLine < len(input) && input[nextLine] != '\n' {
		nextLine++
	}
	if nextLine == len(input) {
		return -1
	} else {
		return nextLine
	}
}

// runs through all restricted tags, and looks for a match
// if matched, returns true, otherwise false
func (pdata *PageMetadata) MatchedTopic(checkTags []string) bool {
	for _, tag := range checkTags {
		if pdata.Topics[tag] == true {
			return true
		}
	}
	return false
}

// returns all the tags within a list as an array of strings
func (pdata *PageMetadata) ListMeta() (
	topics []string, keywords []string, authors []string) {
	for oneTag, _ := range pdata.Topics {
		topics = append(topics[:], oneTag)
	}
	sort.Strings(topics)

	for oneKeyword, _ := range pdata.Keywords {
		keywords = append(keywords[:], oneKeyword)
	}
	sort.Strings(keywords)

	for oneAuthor, _ := range pdata.Authors {
		authors = append(authors[:], oneAuthor)
	}
	sort.Strings(topics)
	return
}

func bodyParseMarkdown(input []byte) []byte {
	// set up the HTML renderer
	renderer := blackfriday.HtmlRenderer(bodyHtmlFlags, "", "")
	return blackfriday.Markdown(input, renderer, bodyExtensions)
}

func tocParseMarkdown(input []byte) []byte {
	// set up the HTML renderer
	renderer := tocRenderer.HtmlRenderer(tocHtmlFlags, "", "")
	return blackfriday.Markdown(input, renderer, tocExtensions)
}
