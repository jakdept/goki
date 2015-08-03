// this file contains a pre-processor to pull some stuff out of the markdown file before parsing it

package gnosis

import (
	"bufio"
	"bytes"
	"errors"
	"html/template"
	"io"
	"sort"
	//"strings"
	//"log"
	"github.com/JackKnifed/blackfriday"
	"os"
)

type PageMetadata struct {
	Keywords  map[string]bool
	Topics    map[string]bool
	Author    map[string]bool
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

// takes a single line of input and determines if it's a top level markdown header
func (pdata *PageMetadata) lineIsTitle(line []byte) bool {
	// trim any whitespace from the start and the end of the line
	line = bytes.TrimSpace(line)

	// run through all of the ='s - make sure they're all correct
	for i := 0; i < len(line); i++ {
		if line[i] != '=' {
			return false
		}
	}

	// if you got here, it should all be legit
	return true
}

// take a given line, and check it against every possible type of tag
func (pdata *PageMetadata) processMetadata(line []byte) {
	pdata.checkMatch(line, []byte("tag"), &pdata.Topics)
	pdata.checkMatch(line, []byte("topic"), &pdata.Topics)
	pdata.checkMatch(line, []byte("category"), &pdata.Topics)

	pdata.checkMatch(line, []byte("keyword"), &pdata.Keywords)
	pdata.checkMatch(line, []byte("meta"), &pdata.Keywords)

	pdata.checkMatch(line, []byte("author"), &pdata.Author)
	pdata.checkMatch(line, []byte("maintainer"), &pdata.Author)
}

func (pdata *PageMetadata) checkMatch(input []byte, looking []byte, tracker *map[string]bool) {
	// trim off any blank spaces at the start of the line
	value := bytes.TrimSpace(input)

	// should be a substring match based on the start of the array
	if len(input) > len(looking) && bytes.Equal(input[:len(looking)], looking) {

		// trim off the target from the []byte
		value = input[len(looking):]

		// trim spaces at the start and at the end
		value = bytes.TrimSpace(value)

		if value[0] == ':' || value[0] == '=' {
			value = bytes.Trim(value, " \t\n=:")
		}

		// replace any spaces in the middle with -'s
		value = bytes.Replace(value, []byte(" "), []byte("-"), -1)

		// suppress any double dashes
		for i := 0; i < len(value)-1; i++ {
			if value[i] == '-' && value[i+1] == '-' {
				value = append(value[:i], value[i+1:]...)
			}
		}

		// now just add the value to the array that you're tracking
		if *tracker != nil {
			(*tracker)[string(value)] = true
		} else {
			*tracker = map[string]bool{string(value): true}
		}
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

	// read a line
	lineBuffer, err := reader.ReadBytes('\n')

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
			lineBuffer = lineBuffer[bytesDone+1:]
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
	nextLine := pdata.findNextLine(input)
	nextLineContent := input[nextLine+1:]
	nextLineContent = bytes.TrimSpace(nextLineContent)

	// step thru each position in the second line
	// if they are all '=', then this whole input is title
	for i:=0; i<len(nextLineContent); i++ {
		if nextLineContent[i] != '=' {
			i = 0
			break 
		} else if i +1 == len(nextLineContent){
			pdata.Title = string(input[:nextLine])
			return len(input)
		}
	}

	pdata.processMetadata(input[:nextLine])

	// reworked header stuff from blackfriday
	for i:=nextLine; i < 6 && i + 1 < len(input); i++{
		if input[i] == '#' && input[i+1] != '#' {
			newTitle := bytes.TrimSpace(input[nextLine:])
			newTitle = bytes.TrimSuffix(newTitle, []byte("#"))
			newTitle = bytes.TrimSpace(newTitle)
			pdata.Title = string(newTitle)
			return len(input)	
		} else if input[i] != ' ' && input[i] != '\t' {
			break
		}
	}
	return nextLine
}

// given input, find where the next line starts
func (pdata *PageMetadata) findNextLine(input []byte) int {
	nextLine := 0
	for nextLine < len(input) && nextLine != '\n' {
		nextLine++
	}
	if nextLine + 1 == len(input) {
		return 0
	} else {
		return nextLine
	}
}

// runs through all restricted tags, and looks for a match
// if matched, returns true, otherwise false
func (pdata *PageMetadata) MatchedTag(checkTags []string) bool {
	for _, tag := range checkTags {
		if pdata.Topics[tag] == true {
			return true
		}
	}
	return false
}

// returns all the tags within a list as an array of strings
func (pdata *PageMetadata) ListMeta() (topics []string, keywords []string) {
	for oneTag, _ := range pdata.Topics {
		topics = append(topics[:], oneTag)
	}
	sort.Strings(topics)

	for oneKeyword, _ := range pdata.Keywords {
		keywords = append(keywords[:], oneKeyword)
	}
	sort.Strings(keywords)
	return
}

// return the bytes to display the tags on the Metadata
// takes the prefix for the tags
func (pdata *PageMetadata) PrintTopics(tagPrefix string) template.HTML {
	response := []byte{}
	openingTag := []byte("<div class='tag'>")
	closingTag := []byte("</div>")
	allTopics, _ := pdata.ListMeta()
	for _, oneTopic := range allTopics {
		response = bytes.Join([][]byte{openingTag, []byte(tagPrefix), []byte(oneTopic), closingTag}, []byte(""))
	}
	return template.HTML(response)
}

// returns the bytes to add the keywrods to the html output
func (pdata *PageMetadata) PrintKeywords() template.HTML {
	if len(pdata.Keywords) > 0 {
		var response []byte
		_, allKeywords := pdata.ListMeta()
		for _, oneKeyword := range allKeywords {
			response = bytes.Join([][]byte{response, []byte(oneKeyword)}, []byte(","))
		}
		// clean up the end of the string and add the ending tag
		response = bytes.Trim(response, ",")
		response = bytes.Join([][]byte{[]byte("<meta name='keywords' content='"), response, []byte("'>")}, []byte(""))

		return template.HTML(response)
	} else {
		return template.HTML("")
	}
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
