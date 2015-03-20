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

type WikiPage struct {
	Title string
	//Body  string
	Body template.HTML
}

var templates = template.Must(template.ParseFiles("wiki.html"))

var wikiFilter = regexp.MustCompile("^/([a-zA-Z0-9_ /]+/)?([a-zA-Z0-9_ ]+)$")
var fileFIlter = regexp.MustCompile("^/([a-zA-Z0-9_ /]+/)?([a-zA-Z0-9_ ]+)?\\.([a-zA-Z0-9_ ]+))?")

func markdownHandler(responsePipe http.ResponseWriter, request *http.Request) {
	var err error

	filteredRequest := wikiFilter.FindStringSubmatch(request.URL.Path)
	config := GetConfig()

	if filteredRequest == nil {
		log.Printf("null request improperly routed to wiki handler %s", request.URL.Path, config.mainserver.prefix)
		http.Error(responsePipe, "Request not allowed", 403)
	}

	if filteredRequest[1] != config.mainserver.prefix {
		log.Printf("request %s was improperly routed to wiki handler %s", request.URL.Path, config.mainserver.prefix)
		http.Error(responsePipe, err.Error(), 500)
	}

	contents, err := ioutil.ReadFile(config.mainserver.prefix + filteredRequest[2] + ".md")
	if err != nil {
		log.Printf("request %s points to an bad file target sent to server %s", request.URL.Path, config.mainserver.prefix)
		http.Error(responsePipe, err.Error(), 403)
	}
	// parse any markdown in the input
	body := template.HTML(blackfriday.MarkdownCommon(contents))

	response := WikiPage{Title: filteredRequest[2], Body: body}
	err = templates.ExecuteTemplate(responsePipe, "wiki.html", response)
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
}

func rawHandler(responsePipe http.ResponseWriter, request *http.Request) {
	var err error

	filteredRequest := wikiFilter.FindStringSubmatch(request.URL.Path)
	config := GetConfig()

	if filteredRequest == nil {
		log.Printf("null request improperly routed to wiki handler %s", request.URL.Path, config.mainserver.prefix)
		http.Error(responsePipe, "Request not allowed", 403)
	}

	if filteredRequest[1] != config.mainserver.prefix {
		log.Printf("request %s was improperly routed to wiki handler %s", request.URL.Path, config.mainserver.prefix)
		http.Error(responsePipe, err.Error(), 500)
	}

	contents, err := ioutil.ReadFile(config.mainserver.prefix + filteredRequest[2] + ".md")
	if err != nil {
		log.Printf("request %s points to an bad file target sent to server %s", request.URL.Path, config.mainserver.prefix)
		http.Error(responsePipe, err.Error(), 403)
	}

	_, err = responsePipe.Write([]byte(contents))
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
}
