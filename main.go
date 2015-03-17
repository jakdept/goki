package main

import (
	"html/template"
	"io/ioutil"
	"regexp"
	"log"
	"sync"
	"encoding/json"

	"net/http"

	"flag"
	//"github.com/FogCreek/mini"

	"github.com/JackKnifed/blackfriday"
	//"github.com/russross/blackfriday"
)

type WikiPage struct {
	Title string
	//Body  stringo
	Body template.HTML
}

var templates = template.Must(template.ParseFiles("wiki.html"))

var validWiki = regexp.MustCompile("^/([a-zA-Z0-9_ ]+)$")

var	configFile = flag.String("config", "", "specify a configuration file")

var staticConfig *configStruct
var configLock = new(sync.RWMutex)

//var validFiles = regexp.MustCompile("^/raw/([a-zA-Z0-9]+)\\.(jpg|gif|jpeg|md|png)$")

//var validSupport = regexp.MustCompile("^/global/([a-zA-Z0-9]+)\\.(css|js)$")
type globalConfigStruct struct{
	port string
	hostname string
}

type serverConfigStruct struct {
	path string
	prefix string
	defaultPage string
	serverType string
	restricted []string
}

type configStruct struct {
	global globalConfigStruct
	mainserver serverConfigStruct
	server []serverConfigStruct
}

func GetConfig() *configStruct {
	configLock.RLock()
	defer configLock.RUnlock()
	return staticConfig
}

func LoadConfig(configFile string) bool {

	if configFile == "" {
		log.Println("no configuration file specified, using ./config.json")
		// return an empty config file
		configFile = "config.json"
	}
	//temp, err := mini.LoadConfiguration(configFile)

	//var fileContents []byte
	//var err error
	// have to read in the line into a byte[] array
	fileContents, err := ioutil.ReadFile(configFile); if err != nil {
		log.Println("Problem loading config file: ", err)
	}

	// UnMarshal the config file that was read in
	temp := new(configStruct)
	err = json.Unmarshal(fileContents, temp)

	//Make sure you were able to read it in
	if err != nil {
		log.Println("parse config error: ", err)
		return false
	}

	configLock.Lock()
	staticConfig = temp
	configLock.Unlock()

	return true
}

func markdownHandler(responsePipe http.ResponseWriter, request *http.Request) {
	validRequest := validWiki.FindStringSubmatch(request.URL.Path)
	config := GetConfig()
	if validRequest == nil {
		http.Error(responsePipe, "Request not allowed", 403)
	}
	contents, err := loadFile(config.mainserver.prefix + validRequest[1] + ".md")
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

func loadFile(filename string) ([]byte, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func main() {
	flag.Parse()

	LoadConfig(*configFile)

	config := GetConfig()

	rawFiles := http.FileServer(http.Dir(config.mainserver.path))
	siteFiles := http.FileServer(http.Dir(config.mainserver.path))

	http.Handle("/raw/", http.StripPrefix("/raw/", rawFiles))
	http.Handle("/site/", http.StripPrefix("/site/", siteFiles))
	http.HandleFunc("/", markdownHandler)

	log.Println(http.ListenAndServe(":"+config.global.port, nil))
}
