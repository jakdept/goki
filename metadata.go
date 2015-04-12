// this file contains a pre-processor to pull some stuff out of the markdown file before parsing it

package gnosis

import (
	"bufio"
	"bytes"
	"errors"
	"html/template"
	"os"
)

type PageMetadata struct {
	Keywords map[string]bool
	Topics   map[string]bool
	Loaded   bool
	Page     []byte
}

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
	pdata.checkMatch(line, []byte("tag"), pdata.Topics)
	pdata.checkMatch(line, []byte("topic"), pdata.Topics)
	pdata.checkMatch(line, []byte("category"), pdata.Topics)

	pdata.checkMatch(line, []byte("keyword"), pdata.Keywords)
	pdata.checkMatch(line, []byte("keywords"), pdata.Keywords)
	pdata.checkMatch(line, []byte("meta"), pdata.Keywords)
}

func (pdata *PageMetadata) LoadPage(pageName string) error {
	f, err := os.Open(pageName)
	reader := bufio.NewReader(f)
	upperLine, fullLine, err := reader.ReadLine()

	// inspect the first line you read
	if err != nil {
		return err
	} else if !fullLine {
		return errors.New("first line I read wasn't a full line")
	} else if pdata.lineIsTitle(upperLine) {
		return errors.New("first line looks an awful lot like the underside of the title o.O")
	}

	lowerLine, fullLine, err := reader.ReadLine()

	// inspect the lower line
	if err != nil {
		return err
	} else if !fullLine {
		return errors.New("second line I read wasn't a full line")
	} else if pdata.lineIsTitle(lowerLine) {

		// read the rest of the page
		var restOfPage []byte
		_, err = reader.Read(restOfPage)
		if err != nil {
			return err
		}
		// if the second line is a title, read the rest of the page in
		// you don't have any metadata to work with here, move on
		pdata.Page = bytes.Join([][]byte{upperLine, lowerLine, restOfPage}, []byte("\n"))

		// you've successfully loaded the page - so return nothing
		pdata.Loaded = true
		return nil
	}

	// if you're at this point, the first line is metadata
	// you gotta process it and work with the next line
	// so let's just read through the file until we hit the title
	for !pdata.lineIsTitle(lowerLine) {
		// process the line
		pdata.processMetadata(upperLine)
		// shift the lower line up
		upperLine = lowerLine
		// read in a new lower line
		lowerLine, fullLine, err = reader.ReadLine()
		if err != nil {
			return err
		} else if !fullLine {
			return errors.New("I filled my buffer with a line")
		}
	}

	// by this point, I should have read everything in - let's read the rest and just return it
	var restOfPage []byte
	_, err = reader.Read(restOfPage)
	pdata.Page = bytes.Join([][]byte{upperLine, lowerLine, restOfPage}, []byte("\n"))

	return err
}

func (pdata *PageMetadata) checkMatch(input []byte, looking []byte, tracker map[string]bool) {
	// trim off any blank spaces at the start of the line
	value := bytes.Trim(input, " \t")

	// should be a substring match based on the start of the array
	if bytes.Equal(input[:len(looking)], looking) {
		// trim off the target from the []byte
		value = input[len(looking):]

		// trim spaces at the start and at the end
		value = bytes.Trim(value, " \t\n")

		if value[0] == ':' || value[0] == '=' {
			value = bytes.Trim(value, " \t\n=:")
		}

		// replace any spaces in the middle with -'s
		bytes.Replace(value, []byte(" "), []byte("-"), -1)

		// suppress any double dashes
		for i := 0; i < len(value)-1; i++ {
			if value[i] == '-' && value[i+1] == '-' {
				value = append(value[:i], value[i+1:]...)
			}
		}

		// now just add the value to the array that you're tracking
		tracker[string(value)] = true
	}
}

// returns all the tags within a list as an array of strings
func (pdata *PageMetadata) ListMeta() ([]string, []string) {
	topics := []string{}
	for oneTag, _ := range pdata.Topics {
		topics = append(topics[:], oneTag)
	}

	keywords := []string{}
	for oneKeyword, _ := range pdata.Keywords {
		keywords = append(keywords[:], oneKeyword)
	}

	return topics, keywords
}

// return the bytes to display the tags on the page
// takes the prefix for the tags
func (pdata *PageMetadata) PrintTopics(tagPrefix string) template.HTML {
	response := []byte{}
	openingTag := []byte("<div class='tag'>")
	closingTag := []byte("</div>")
	for oneTag, _ := range pdata.Topics {
		response = bytes.Join([][]byte{openingTag, []byte(tagPrefix), []byte(oneTag), closingTag}, []byte(""))
	}
	return template.HTML(response)
}

// returns the bytes to add the keywrods to the html output
func (pdata *PageMetadata) PrintKeywords() template.HTML {
	response := []byte("<meta name='keywords' content='")
	for oneKeyword, _ := range pdata.Keywords {
		response = bytes.Join([][]byte{response, []byte(oneKeyword)}, []byte(","))
	}
	// clean up the end of the string and add the ending tag
	response = bytes.TrimSuffix(response, []byte{','})
	response = append(response, []byte("'>")...)

	return template.HTML(response)
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
