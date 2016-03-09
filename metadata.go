// this file contains a pre-processor to pull some stuff out of the markdown file before parsing it

package goki

import (
	"bufio"
	"bytes"
	"errors"
	// "html/template"
	"io"
	"sort"
	// "strings"
	// "log"
	"github.com/JackKnifed/blackfriday"
	"os"
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
		blackfriday.HTML_OMIT_CONTENTS

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

// take a given line, and check it against every possible type of tag
func (pdata *PageMetadata) processMetadata(line []byte) {
	pdata.checkMatch(line, []byte("tag"), &pdata.Topics)
	pdata.checkMatch(line, []byte("topic"), &pdata.Topics)
	pdata.checkMatch(line, []byte("category"), &pdata.Topics)

	pdata.checkMatch(line, []byte("keyword"), &pdata.Keywords)
	pdata.checkMatch(line, []byte("meta"), &pdata.Keywords)

	pdata.checkMatch(line, []byte("author"), &pdata.Authors)
	pdata.checkMatch(line, []byte("maintainer"), &pdata.Authors)
}

func (pdata *PageMetadata) checkMatch(input []byte, looking []byte, tracker *map[string]bool) {
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

func (pdata *PageMetadata) readRestOfPage(r *bufio.Reader) error {
	// read the rest of the page
	var restOfPage []byte
	var err error

	for err == nil {
		// read a line, and then add it to pdata
		restOfPage, err = r.ReadBytes('\n')
		pdata.Page = append(pdata.Page, restOfPage...)
	}

	if err == io.EOF {
		return nil
	} else {
		return err
	}
}

func (pdata *PageMetadata) LoadPage(pageName string) error {
	// open the file
	f, err := os.Open(pageName)
	defer f.Close()
	reader := bufio.NewReader(f)
	if err != nil {
		return err
	}
	pdata.FileStats, err = os.Stat(pageName)

	// read a line and sneak a newline on the front
	lineBuffer, err := reader.ReadBytes('\n')
	lineBuffer = append([]byte("\n"), lineBuffer...)

	for err != io.EOF {
		// check the first line you read
		if err != nil {
			return errors.New("error reading from file - " + err.Error())
		}
		bytesDone := pdata.isTitle(lineBuffer)
		if bytesDone == len(lineBuffer) {
			return pdata.readRestOfPage(reader)
		} else {
			var newLine []byte
			lineBuffer = lineBuffer[bytesDone:]
			newLine, err = reader.ReadBytes('\n')
			lineBuffer = append(lineBuffer, newLine...)
		}
	}
	return errors.New("need to hit a title in the file")
}

// determines if the next two lines contain a title line
// if the first line is not a line, treat it as metadata
// return the amount of characters processed if not a new line
// if title line, return total length of the input
func (pdata *PageMetadata) isTitle(input []byte) int {
	newline := []byte("\n")
	nextLine := bytes.Index(input, newline)
	if nextLine == -1 {
		return pdata.isOneLineTitle(input)
	}
	if rv := pdata.isOneLineTitle(input[:nextLine+1]); rv != 0 {
		return rv
	}
	if nextLine >= len(input) {
		return 0
	}

	lineAfter := bytes.Index(input[nextLine+1:], newline)
	if lineAfter == -1 {
		lineAfter = len(input) - 1
	} else {
		lineAfter += nextLine + 1
	}

	if rv := pdata.isTwoLineTitle(input[:lineAfter+1]); rv != 0 {
		return rv
	}

	pdata.processMetadata(input[:nextLine])
	if rv := pdata.isOneLineTitle(input[nextLine+1 : lineAfter+1]); rv != 0 {
		return rv + nextLine + 1
	} else {
		return nextLine + 1
	}
}

// checks to see if the first lines of a []byte contain a markdown title
// returns the number of characters to lose
// 0 indicates failure (no characters to lose)
func (pdata *PageMetadata) isOneLineTitle(input []byte) int {
	var singleLine []byte
	var endOfLine int

	if endOfLine = pdata.findNextLine(input); endOfLine != -1 {
		singleLine = bytes.TrimSpace(input[:endOfLine])
	} else {
		endOfLine = len(input) - 1
		singleLine = bytes.TrimSpace(input)
	}

	if len(singleLine) > 2 && singleLine[0] == '#' && singleLine[1] != '#' {
		singleLine = bytes.Trim(singleLine, "#")
		pdata.Title = string(bytes.TrimSpace(singleLine))
		return endOfLine + 1
	}
	return 0
}

// checks to see if the first two lines of a []byte contain a markdown title
// returns the number of characters to lose
// 0 indicates failure (no characters to lose)
func (pdata *PageMetadata) isTwoLineTitle(input []byte) int {
	var firstNewLine, secondNewLine int

	if firstNewLine = pdata.findNextLine(input); firstNewLine == -1 {
		return 0
	}
	secondNewLine = pdata.findNextLine(input[firstNewLine+1:])
	if secondNewLine == -1 {
		secondNewLine = len(input) - 1
	} else {
		secondNewLine += firstNewLine + 1
	}

	secondLine := bytes.TrimSpace(input[firstNewLine+1 : secondNewLine+1])
	if len(secondLine) >= 2 {
		secondLine = bytes.TrimLeft(secondLine, "=")
		if len(secondLine) == 0 {
			pdata.Title = string(bytes.TrimSpace(input[:firstNewLine]))
			return secondNewLine + 1
		}
	}
	return 0
}

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
func (pdata *PageMetadata) ListMeta() (topics []string, keywords []string, authors []string) {
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
	renderer := blackfriday.HtmlRenderer(tocHtmlFlags, "", "")
	return blackfriday.Markdown(input, renderer, tocExtensions)
}
