// this file contains a pre-processor to pull some stuff out of the markdown file before parsing it

package gnosis

import (
	"bufio"
	"bytes"
	"errors"
	"html/template"
	"io"
	//"log"
	"os"
)

type PageMetadata struct {
	Keywords map[string]bool
	Topics   map[string]bool
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
	pdata.checkMatch(line, []byte("tag"), &pdata.Topics)
	pdata.checkMatch(line, []byte("topic"), &pdata.Topics)
	pdata.checkMatch(line, []byte("category"), &pdata.Topics)

	pdata.checkMatch(line, []byte("keyword"), &pdata.Keywords)
	pdata.checkMatch(line, []byte("keywords"), &pdata.Keywords)
	pdata.checkMatch(line, []byte("meta"), &pdata.Keywords)
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

func (pdata *PageMetadata) readRestOfPage(topLine []byte, bottomLine []byte, r *bufio.Reader) error {
	// read the rest of the page
	var restOfPage []byte
	var err error

	// put the start of stuff into the final destination
	pdata.Page = bytes.Join([][]byte{topLine, bottomLine, []byte("")}, []byte(""))

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
	reader := bufio.NewReader(f)
	if err != nil {
		return err
	}

	// read a line
	upperLine, err := reader.ReadBytes(byte('\n'))

	// check the first line you read
	if err == io.EOF {
		return errors.New("I only read in... one line?")
	} else if err != nil {
		return errors.New("first line error - " + err.Error())
	}

	// read a second line - this might actually be a real line
	lowerLine, err := reader.ReadBytes('\n')
	// inspect the lower line
	if err == io.EOF {
		return errors.New("Is this page just a title?")
	} else if err != nil {
		return errors.New("secont line error - " + err.Error())
	} else if pdata.lineIsTitle(lowerLine) {
		return pdata.readRestOfPage(upperLine, lowerLine, reader)
	}

	// if you're at this point, the first line is metadata
	// you gotta process it and work with the next line
	for !pdata.lineIsTitle(lowerLine) && err != io.EOF {
		// process the line
		pdata.processMetadata(upperLine)
		// shift the lower line up
		upperLine = lowerLine
		// read in a new lower line
		lowerLine, err = reader.ReadBytes('\n')
		if err == io.EOF {
			return errors.New("never hit a title")
		} else if err != nil {
			return err
		}
	}

	// by this point, I should have read everything in - let's read the rest and just return it
	return pdata.readRestOfPage(upperLine, lowerLine, reader)
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
