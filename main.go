package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"regexp"

	"net/http"

	"flag"
	//"github.com/FogCreek/mini"

	"github.com/JackKnifed/blackfriday"
	"github.com/JackKnifed/gnosis"
	//"github.com/russross/blackfriday"
)

type WikiPage struct {
	Title string
	//Body  stringo
	Body template.HTML
}

var templates = template.Must(template.ParseFiles("wiki.html"))

var validWiki = regexp.MustCompile("^/([a-zA-Z0-9_ ]+)$")

var configFile = flag.String("config", "", "specify a configuration file")

//var validFiles = regexp.MustCompile("^/raw/([a-zA-Z0-9]+)\\.(jpg|gif|jpeg|md|png)$")

//var validSupport = regexp.MustCompile("^/global/([a-zA-Z0-9]+)\\.(css|js)$")

func markdownHandler(responsePipe http.ResponseWriter, request *http.Request) {
	validRequest := validWiki.FindStringSubmatch(request.URL.Path)
	config := gnosis.GetConfig()
	if validRequest == nil {
		http.Error(responsePipe, "Request not allowed", 403)
	}
	contents, err := ioutil.ReadFile(config.mainserver.prefix + validRequest[1] + ".md")
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
	// parse any markdown in the input
	body := template.HTML(blackfriday.MarkdownCommon(contents))

	response := WikiPage{Title: validRequest[1], Body: body}
	err = templates.ExecuteTemplate(responsePipe, "wiki.html", response)
	if err != nil {
		http.Error(responsePipe, err.Error(), http.StatusInternalServerError)
	}
}

/*
func contentHandler(responsePipe http.ResponseWriter, request *http.Request) {
	validRequest := validFiles.FindStringSubmatch(request.URL.Path)
	if validRequest == nil {
		http.Error(responsePipe, "Request not allowed", 403)
	}
	strippedRequest := http.StripPrefix("/raw/", validRequest[0])
	http.ServeFile(responsePipe, request, strippedRequest)
	http.FileServer()
	contents, err := loadFile(validRequest)
	if err != nil {
		http.Error(responsePipe, err, 500)
	}
	response := contents
	err := templates.ExecuteTemplate(responsePipe, "markdown.html", p)
	if err != nil {
		http.Error(responsePipe, err.Error(), http.StatusInternalServerError)
	}

}
*/

/*
func loadFile(filename string) ([]byte, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return contents, nil
}
*/

func main() {
	flag.Parse()

	gnosis.LoadConfig(*configFile)

	config := gnosis.GetConfig()

	rawFiles := http.FileServer(http.Dir(config.mainserver.path))
	siteFiles := http.FileServer(http.Dir(config.mainserver.path))

	http.Handle("/raw/", http.StripPrefix("/raw/", rawFiles))
	http.Handle("/site/", http.StripPrefix("/site/", siteFiles))
	http.HandleFunc("/", markdownHandler)

	log.Println(http.ListenAndServe(":"+config.global.port, nil))

}
