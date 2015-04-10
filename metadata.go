// this file contains a pre-processor to pull some stuff out of the markdown file before parsing it

package gnosis

import (
	"bufio"
	"bytes"
	"html/template"
	"io"
	"os"
)

type pageMetadata struct {
	Keywords map[string]bool
	Tags     map[string]bool
	Loaded   bool
	Page     []byte
}

var pdata pageMetadata

func (pdata *PageMetadata) lineIsTitle(line []byte) bool {
	finalLength := len(line)
	i := 0

	// if the row doesn't start with tabs, spaces, or ='s
	if (data[i] != ' ' && data[i] != '=') && data[i] != '\t' {
		return false
	}

	// skip any spaces or tabs at the start
	for data[i] == ' ' || data[i] == '\t' {
		i++
	}

	// if the next item's not a =, bail out
	if data[i] != '=' {
		return false
	}

	// run through all of the ='s
	for data[i] == '=' {
		i++
	}

	if data[i] != ' ' && data[i] != '\t' && data[i] != '\n' {
		return false
	}

	//ditch all spaces after any ='s
	for data[i] == ' ' || data[i] == '\t' {
		i++
	}

	if finalLength == i+1 {
		return true
	} else {
		return false
	}
}

func (pdata *PageMetadata) checkMatch(input []byte, looking []byte, tracker []string) {
	// trim off any blank spaces at the start of the line
	value := bytes.Trim(line, " \t")

	if input[:len(looking)] == looking {
		// trim off the target from the []byte
		value = input[len(looking):]

		// trim spaces at the start and at the end
		value = bytes.Trim(value, " \t\n")

		if value[0] == ':' || value[0] == '=' {
			value = bytes.Trim(value, " \t\n=:")
		}

		// replace any spaces in the middle with -'s
		bytes.Replace(value, " ", "-", -1)

		// suppress any double dashes
		for i := 1; i < len(value); i++ {
			if value[i-1] == '-' && value[1] == '-' {
				value = value[:i] + value[i+1:]
			}
		}

		// now just add the value to the array that you're tracking
		tracker[value] = true
	}
}

// returns all the tags within a list as an array of strings
func (pdata *PageMetadata) ListMeta() ([]string, []sting) {
	tags := new([]string)
	for oneTag, _ := range pdata.Tags {
		tags.Append(oneTag)
	}

	keywords := new([]string)
	for oneKeyword, _ := range pdata.Keywords {
		keywords.Append(oneKeyword)
	}

	return tags, keywords
}

// return the bytes to display the tags on the page
// takes the prefix for the tags
func (pdata *PageMetadata) PrintTags(tagPrefix string) template.HTML {
	response := new(string)
	for oneTag, _ := range pdata.Tags {
		response += "<div class='tag'>"
		response += tagPrefix
		response += oneTag
		response += "</div>"
	}
	return template.HTML(response)
}

// returns the bytes to add the keywrods to the html output
func (pdata *PageMetadata) PrintKeywords() template.HTML {
	response := string("<meta name='keywords' content='")
	for oneKeyword, _ := range pdata.Keywords {
		response += oneKeyword
		response += ','
	}
	// clean up the end of the string and add the ending tag
	response = response.TrimSuffix(response, ',')
	response += "'>"

	return template.HTML(response)
}

// runs through all restricted tags, and looks for a match
// if matched, returns true, otherwise false
func (pdata *PageMetadata) MatchedTag(checkTags []string) bool {
	for _, tag := range checkTags {
		if pdata.Tags[tag] == true {
			return true
		}
	}
	return false
}

func (pdata *PageMetadata) ProcessMetadata(line []byte) error {
	pdata.checkMatch(line, "tag", pdata.Tags)
	pdata.checkMatch(line, "topic", pdata.Tags)
	pdata.checkMatch(line, "category", pdata.Tags)

	pdata.checkMatch(line, "keyword", pdata.Keywords)
	pdata.checkMatch(line, "keywords", pdata.Keywords)
	pdata.checkMatch(line, "meta", pdata.Keywords)
}

func (pdata *PageMetadata) LoadPage(pageName string) error {
	f, err := os.Open(pageName)
	reader := bufio.NewReader(f)
	upperLine, fullLine, err := reader.ReadLine()

	// inspect the first line you read
	if err != nil {
		return err
	} else if !fullLine {
		return err.New("first line I read wasn't a full line")
	} else if lineIsTitle(upperLine) {
		return err.New("first line looks an awful lot like the underside of the title o.O")
	}

	lowerLine, fullLine, err := reader.ReadLine()

	// inspect the lower line
	if err != nil {
		return err
	} else if !fullLine {
		return err.New("second line I read wasn't a full line")
	} else if lineIsTitle(lowerLine) {
		// if the second line is a title, read the rest of the page in
		// you don't have any metadata to work with here, move on
		upperLine.Append('\n')
		upperLine.Append(lowerLine)
		upperLine.Append('\n')

		_, err = reader.Read(lowerLine)
		if err != nil {
			return err
		}

		// you've successfully loaded the page - so return nothing
		pdata.Loaded = true
		return nil
	}

	// if you're at this point, the first line is metadata
	// you gotta process it and work with the next line
	// so let's just read through the file until we hit the title
	for !lineIsTitle(lowerLine) {
		// process the line
		ProcessMetadata(upperLine)
		// shift the lower line up
		upperLine = lowerLine
		// read in a new lower line
		lowerLine, fullLine, err := reader.ReadLine()
		if err != nil {
			return err
		} else if !fullLine {
			return err.New("I filled my buffer with a line")
		}
	}

	// by this point, I should have read everything in - let's read the rest and just return it
	upperLine.Append('\n')
	upperLine.Append(lowerLine)
	upperLine.Append('\n')

	_, err = reader.Read(lowerLine)
	return err
}
