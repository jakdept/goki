package gnosis

// package file will contain MarkdownHandler and RawHandler to handle incoming requests
// Whole thing needs to be written

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/JackKnifed/blackfriday"
)

var fileFIlter = regexp.MustCompile("^(/([a-zA-Z0-9_ /]+/)?)([a-zA-Z0-9_ ]+)?(\\.)?([a-zA-Z0-9_ ]+)?")

func RawHandler(responsePipe http.ResponseWriter, request *http.Request) {
	var err error

	filteredRequest := wikiFilter.FindStringSubmatch(request.URL.Path)
	config := GetConfig()

	if filteredRequest == nil {
		log.Printf("null request improperly routed to wiki handler %s", request.URL.Path, config.Mainserver.Prefix)
		http.Error(responsePipe, "Request not allowed", 403)
	}

	if filteredRequest[1] != config.Mainserver.Prefix {
		log.Printf("request %s was improperly routed to wiki handler %s", request.URL.Path, config.Mainserver.Prefix)
		http.Error(responsePipe, err.Error(), 500)
	}

	contents, err := ioutil.ReadFile(config.Mainserver.Prefix + filteredRequest[2] + ".md")
	if err != nil {
		log.Printf("request %s points to an bad file target sent to server %s", request.URL.Path, config.Mainserver.Prefix)
		http.Error(responsePipe, err.Error(), 403)
	}

	_, err = responsePipe.Write([]byte(contents))
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
}

const (
	bodyHtmlFlags = 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES

	bodyExtensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_AUTO_HEADER_IDS |
		blackfriday.EXTENSION_TITLEBLOCK

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

type WikiPage struct {
	Title string
	ToC   template.HTML
	Body  template.HTML
}

var templates = template.Must(template.ParseFiles("wiki.html"))

var wikiFilter = regexp.MustCompile("^(/([a-zA-Z0-9_ /]+/)?)([a-zA-Z0-9_ ]+)$")

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

func MarkdownHandler(responsePipe http.ResponseWriter, request *http.Request) {
	var err error

	filteredRequest := wikiFilter.FindStringSubmatch(request.URL.Path)
	config := GetConfig()

	if filteredRequest == nil {
		log.Printf("null request [ %s ] improperly routed to wiki handler [ %s ]", request.URL.Path, config.Mainserver.Prefix)
		http.Error(responsePipe, "Request not allowed", 403)
	} else {
		if filteredRequest[1] != config.Mainserver.Prefix {
			log.Printf("request %s was improperly routed to wiki handler %s", request.URL.Path, config.Mainserver.Prefix)
			http.Error(responsePipe, err.Error(), 500)
		}

		contents, err := ioutil.ReadFile(config.Mainserver.Path + filteredRequest[3] + ".md")
		if err != nil {
			log.Printf("request [ %s ] points to an bad file target [ %s ]sent to server %s", request.URL.Path, filteredRequest[3], config.Mainserver.Prefix)
			http.Error(responsePipe, err.Error(), 403)
		}
		// parse any markdown in the input
		body := template.HTML(bodyParseMarkdown(contents))

		toc := template.HTML(tocParseMarkdown(contents))

		response := WikiPage{Title: filteredRequest[3], ToC: toc, Body: body}
		err = templates.ExecuteTemplate(responsePipe, "wiki.html", response)
		if err != nil {
			http.Error(responsePipe, err.Error(), 500)
		}
	}
}
