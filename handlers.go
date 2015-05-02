package gnosis


// package file will contain MarkdownHandler and RawHandler to handle incoming requests
// Whole thing needs to be written

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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

type Page struct {
	Title    string
	ToC      template.HTML
	Body     template.HTML
	Topics   template.HTML
	Keywords template.HTML
}

func stripRequestRouting(stripPath string, request *http.Request) (*http.Request, error) {
	if string(request.URL.Path[0]) != "/" {
		err := errors.New("not compatible with relative requests")
		return nil, err
	}
	lastChar := len(stripPath) - 1
	if string(stripPath[lastChar:]) != "/" {
		err := errors.New("passed a request route that does not end in a /")
		return nil, err
	}
	if string(stripPath[0]) != "/" {
		err := errors.New("passed a request route that does not start in a /")
		return nil, err
	}
	if len(stripPath) > len(request.URL.Path) {
		err := errors.New("request routing path longer than request path")
		return nil, err
	}
	if stripPath != string(request.URL.Path[:len(stripPath)]) {
		err := errors.New("request does not match up to the routed path")
		return nil, err
	}

	returnRequest := request
	returnRequest.URL.Path = string(request.URL.Path[len(stripPath)-1:])
	return returnRequest, nil
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

// Of note - this markdown handler is not a direct handler
func MarkdownHandler(responsePipe http.ResponseWriter, rawRequest *http.Request, serverConfig ServerSection) {

	var err error

	// break up the request parameters - for reference, regex is listed below
	//filteredRequest, err := wikiFilter.FindStringSubmatch(request.URL.Path)

	request, err := stripRequestRouting(serverConfig.Prefix, rawRequest)
	if err != nil {
		log.Printf("request [ %s ] was passed to the wrong handler - got %v", request.URL.Path, err)
		http.Error(responsePipe, "Request not allowed", 403)
		return
	}

	// If the request is empty, set it to the default.
	if request.URL.Path == "" || request.URL.Path == "/" {
		request.URL.Path = serverConfig.DefaultPage
	}

	// If the request doesn't end in .md, add that
	if !strings.HasSuffix(request.URL.Path, ".md") {
		request.URL.Path = request.URL.Path + ".md"
	}

	pdata := new(PageMetadata)
	err = pdata.LoadPage(serverConfig.Path + request.URL.Path)
	if err != nil {
		log.Printf("request [ %s ] points to an bad file target sent to server %s", request.URL.Path, serverConfig.Prefix)
		http.Error(responsePipe, err.Error(), 404)
		return
	}

	if pdata.MatchedTag(serverConfig.Restricted) {
		log.Printf("request [ %s ] was against a page with a restricted tag", request.URL.Path)
		http.Error(responsePipe, err.Error(), 403)
		return
	}

	// parse any markdown in the input
	body := template.HTML(bodyParseMarkdown(pdata.Page))
	toc := template.HTML(tocParseMarkdown(pdata.Page))
	keywords := pdata.PrintKeywords()
	// ##TODO## need to move the topic URL to the config
	topics := pdata.PrintTopics(serverConfig.TopicURL)

	// ##TODO## before you can use a template, you have to get the template lock to make sure you don't mess with someone else reading it
	response := Page{Title: "", ToC: toc, Body: body, Keywords: keywords, Topics: topics}
	err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, response)
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
}

func FindExtension(s string) (string, error) {
	for i := len(s); i > 0; i-- {
		if string(s[i]) == "." {
			return s[i:], nil
		}
	}
	return "", errors.New("found no extension")
}

func RawHandler(responsePipe http.ResponseWriter, rawRequest *http.Request, serverConfig ServerSection) {

	var err error

	request, err := stripRequestRouting(serverConfig.Prefix, rawRequest)
	if err != nil {
		log.Printf("request [ %s ] was passed to the wrong handler - got %v", request.URL.Path, err)
		http.Error(responsePipe, "Request not allowed", 403)
		return
	}

	// If the request is empty, set it to the default.
	if request.URL.Path == "" || request.URL.Path == "/" {
		request.URL.Path = serverConfig.DefaultPage
	}

	// If the request is a blocked restriction, shut it down.
	//extension, err := FindExtension(request.URL.Path)
	for _, restricted := range serverConfig.Restricted {
		if strings.HasSuffix(request.URL.Path, restricted) {
			log.Printf("request %s was improperly routed to the file handler with an disallowed extension %s", request.URL.Path, restricted)
			http.Error(responsePipe, "Request not allowed", 403)
			return
		}
	}

	// Load the file - 404 on failure.
	contents, err := ioutil.ReadFile(serverConfig.Path + request.URL.Path)
	if err != nil {
		log.Printf("request [ %s ] points to an bad file target sent to server %s - %v", request.URL.Path, serverConfig.Prefix, err)
		http.Error(responsePipe, err.Error(), 404)
		return
	}

	_, err = responsePipe.Write([]byte(contents))
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
	return
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
