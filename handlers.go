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
	Title    string
	ToC      template.HTML
	Body     template.HTML
	Topics   template.HTML
	Keywords template.HTML
}

var wikiFilter = regexp.MustCompile("^(/([a-zA-Z0-9_ /]+/)?)([a-zA-Z0-9_ ]+)(.md)?$")

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

// Of note - this markdown handler is not a direct handler
func MarkdownHandler(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error

	// break up the request parameters - for reference, regex is listed below
	filteredRequest := wikiFilter.FindStringSubmatch(request.URL.Path)

	// if there are no matches, the regex ovbiously didn't match up
	if filteredRequest == nil {
		log.Printf("null request [ %s ] improperly routed to wiki handler [ %s ]", request.URL.Path, serverConfig.Prefix)
		http.Error(responsePipe, "Request not allowed", 403)
	} else {
		if filteredRequest[1] != serverConfig.Prefix {
			log.Printf("request %s was improperly routed to wiki handler %s", request.URL.Path, serverConfig.Prefix)
			http.Error(responsePipe, err.Error(), 500)
		}

		if filteredRequest[2] == "" && filteredRequest[3] == "" {
			filteredRequest[2] = serverConfig.DefaultPage
		}

		pdata := new(PageMetadata)
		err = pdata.LoadPage(serverConfig.Path + filteredRequest[3] + ".md")
		// ## TODO ## need to add better error reporting for pages
		if err != nil {
			log.Printf("request [ %s ] points to an bad file target [ %s ]sent to server %s", request.URL.Path, filteredRequest[3], serverConfig.Prefix)
			http.Error(responsePipe, err.Error(), 404)
		}

		if pdata.MatchedTag(serverConfig.Restricted) {
			log.Printf("request [ %s ] was against a page with a restricted tag", request.URL.Path)
			http.Error(responsePipe, err.Error(), 403)
		}

		// parse any markdown in the input
		body := template.HTML(bodyParseMarkdown(pdata.Page))
		toc := template.HTML(tocParseMarkdown(pdata.Page))
		keywords := pdata.PrintKeywords()
		// ##TODO## need to move the topic URL to the config
		topics := pdata.PrintTopics("/topic/")

		// ##TODO## before you can use a template, you have to get the template lock to make sure you don't mess with someone else reading it
		response := WikiPage{Title: filteredRequest[3], ToC: toc, Body: body, Keywords: keywords, Topics: topics}
		err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, response)
		if err != nil {
			http.Error(responsePipe, err.Error(), 500)
		}
	}
}

var fileFilter = regexp.MustCompile("^(/([a-zA-Z0-9_ /]+/)?)([a-zA-Z0-9_ ]+)?(\\.)?([a-zA-Z0-9_ ]+)?")

func RawHandler(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error

	// break up the request parameters - for reference, regex is listed below
	filteredRequest := fileFilter.FindStringSubmatch(request.URL.Path)

	// if there are no matches, the regex ovbiously didn't match up
	if filteredRequest == nil {
		log.Printf("null request [ %s ] improperly routed to file handler [ %s ]", request.URL.Path, serverConfig.Prefix)
		http.Error(responsePipe, "Request not allowed", 403)
	} else {
		if filteredRequest[1] != serverConfig.Prefix {
			log.Printf("request %s was improperly routed to file handler %s", request.URL.Path, serverConfig.Prefix)
			http.Error(responsePipe, err.Error(), 500)
		}

		if filteredRequest[2] == "" && filteredRequest[3] == "" {
			filteredRequest[2] = serverConfig.DefaultPage
		}

		for _, restricted := range serverConfig.Restricted {
			if restricted == filteredRequest[4] {
				log.Printf("request %s was improperly routed to the file handler with an disallowed extension %s", request.URL.Path, filteredRequest[4])
				http.Error(responsePipe, "Request not allowed", 403)
			}
		}

		contents, err := ioutil.ReadFile(serverConfig.Path + filteredRequest[3] + ".md")
		if err != nil {
			log.Printf("request [ %s ] points to an bad file target [ %s ]sent to server %s", request.URL.Path, filteredRequest[3], serverConfig.Prefix)
			http.Error(responsePipe, err.Error(), 404)
		}

		_, err = responsePipe.Write([]byte(contents))
		if err != nil {
			http.Error(responsePipe, err.Error(), 500)
		}
	}
}

func MakeHandler(handlerConfig ServerSection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch handlerConfig.ServerType {
		case "markdown":
			MarkdownHandler(w, r, handlerConfig)
		case "raw":
			RawHandler(w, r, handlerConfig)
		default:
			log.Printf("Bad server type [%s]", handlerConfig.ServerType)
		}
	}
}
