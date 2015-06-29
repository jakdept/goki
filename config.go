package gnosis

import (
	"encoding/json"
	"net/http"
	"html/template"
	"io/ioutil"
	"log"
	"sync"
)

var staticConfig *Config
var configLock = new(sync.RWMutex)

type Config struct {
	Global      GlobalSection
	Redirects   []RedirectSection
	Server      []ServerSection
	Indexes     []IndexSection
	TemplateDir string
}

//var templates = template.Must(template.ParseFiles("/var/wiki-backend/wiki.html"))
//template.New("allTemplates")
var allTemplates = template.New("allTemplates")
var templateLock = new(sync.RWMutex)

type GlobalSection struct {
	Port        string
	Hostname    string
	TemplateDir string
}

type RedirectSection struct {
	Requested string
	Target    string
	Code      int
}

type IndexSection struct {
	WatchDirs      map[string]string // Location that we will be watching for updates
	WatchExtension string   // file extensions that we will watch for within that dir
	IndexPath      string   //location to store the index
	IndexType      string   // type of index - likely "en"
	IndexName      string   // name of the index
	Restricted     []string // Tags to restrict indexing on
}

type ServerSection struct {
	Path        string   // filesystem path to serve out
	Prefix      string   // Web URL Prefix - alternatively the prefix for a search handler
	Default	string   // Default page to serve if empty URI
	Template    string   // Template file to build the web URL from
	ServerType  string   // markdown, raw, or search to denote the type of Server handle
	TopicURL    string   // URI prefix to redirect to topic pages
	Restricted  []string // list of restricts - extensions for raw, topics for markdown
}

func GetConfig() *Config {
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

	// have to read in the line into a byte[] array
	fileContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("Problem loading config file: %s", err.Error())
	}

	// UnMarshal the config file that was read in
	temp := new(Config)

	err = json.Unmarshal(fileContents, temp)
	//Make sure you were able to read it in
	if err != nil {
		log.Printf("parse config error: %s", err.Error())
		log.Print(temp)
		return false
	}

	configLock.Lock()
	staticConfig = temp
	configLock.Unlock()

	return true
}

func RenderTemplate(responsePipe http.ResponseWriter, templateName string, data interface{}) error {
		templateLock.RLock()
		defer templateLock.RUnlock()

		return allTemplates.ExecuteTemplate(responsePipe, templateName, data)
}

func ParseTemplates(globalConfig GlobalSection) {
	var err error
	//newTemplate := template.New("newTemplate")
	newTemplate, err := template.ParseGlob(globalConfig.TemplateDir + "*")
	if err != nil {
		log.Println("Found an invalid template, abandoning updating templates")
		return
	}

	loadedTemplates := newTemplate.Templates()
	for _, individualTemplate := range loadedTemplates {
		log.Printf("Loaded template %s ", individualTemplate.Name())
	}

	if err == nil {
		templateLock.Lock()
		defer templateLock.Unlock()
		allTemplates = newTemplate
	}

}
