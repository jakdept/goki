package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"sync"
)

var staticConfig *GlobalSection
var configLock = new(sync.RWMutex)

//var templates = template.Must(template.ParseFiles("/var/wiki-backend/wiki.html"))
//template.New("allTemplates")
var allTemplates = template.New("allTemplates")
var templateLock = new(sync.RWMutex)

type GlobalSection struct {
	Address     string
	Port        string
	Hostname    string
	TemplateDir string
	CertFile    string
	KeyFile     string
	Indexes     []IndexSection
	Redirects   []RedirectSection
}

type RedirectSection struct {
	Requested string
	Target    string
	Code      int
}

type IndexSection struct {
	WatchDirs      map[string]string // physical -> URI Location that we will be watching for updates
	WatchExtension string            // file extensions that we will watch for within that dir
	IndexPath      string            //location to store the index
	IndexType      string            // type of index - likely "en"
	IndexName      string            // name of the index
	Restricted     []string          // Tags to restrict indexing on
	Handlers       []ServerSection
}

type ServerSection struct {
	Path             string   // filesystem path to serve out
	Prefix           string   // Web URL Prefix - alternatively the prefix for a search handler
	Default          string   // Default page to serve if empty URI - alternatively the facet to list against
	Template         string   // Template file to build the response from
	FallbackTemplate string   // template to fall back to for each handlers
	ServerType       string   // markdown, raw, search, or facet to denote the type of Server handle
	TopicURL         string   // URI prefix to redirect to topic pages
	Restricted       []string // list of restricts - extensions for raw, topics for markdown
}

func GetConfig() *GlobalSection {
	configLock.RLock()
	defer configLock.RUnlock()
	return staticConfig
}

func LoadConfig(configFile string) error {
	if configFile == "" {
		// log.Println("no configuration file specified, using ./config.json")
		// return an empty config file
		configFile = "config.json"
	}

	// have to read in the line into a byte[] array
	fileContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return &Error{Code: ErrReadConfig, innerError: err, value: configFile}
	}

	// UnMarshal the config file that was read in
	temp := new(GlobalSection)

	err = json.Unmarshal(fileContents, temp)
	//Make sure you were able to read it in
	if err != nil {
		return &Error{Code: ErrParseConfig, value: temp, innerError: err}
	}

	CleanConfig(temp)

	configLock.Lock()
	staticConfig = temp
	configLock.Unlock()

	return nil
}

// ## TODO ##
func CleanConfig(config *GlobalSection) {
	config.TemplateDir = filepath.Clean(config.TemplateDir)
	for r := range config.Redirects {
		config.Redirects[r].Requested = path.Clean(config.Redirects[r].Requested)
		config.Redirects[r].Target = path.Clean(config.Redirects[r].Target)
		if config.Redirects[r].Code == 0 {
			config.Redirects[r].Code = 301
		}
	}

	for _, indexSection := range config.Indexes {
		for origDirPath, origWebPath := range indexSection.WatchDirs {
			delete(indexSection.WatchDirs, origDirPath)
			indexSection.WatchDirs[filepath.Clean(origDirPath)] = path.Clean(origWebPath)
		}

		for i := range indexSection.Handlers {
			indexSection.Handlers[i].Path = filepath.Clean(indexSection.Handlers[i].Path)
			indexSection.Handlers[i].Prefix = path.Clean(indexSection.Handlers[i].Prefix)
			indexSection.Handlers[i].Default = path.Clean(indexSection.Handlers[i].Default)
			indexSection.Handlers[i].TopicURL = path.Clean(indexSection.Handlers[i].TopicURL)
		}
	}
}

func RenderTemplate(responsePipe http.ResponseWriter, templateName string,
	data interface{}) error {
	templateLock.RLock()
	defer templateLock.RUnlock()

	return allTemplates.ExecuteTemplate(responsePipe, templateName, data)
}

func ParseTemplates(globalConfig GlobalSection) error {
	// log.Printf("Parsing templates in [%q]", globalConfig.TemplateDir)
	newTemplate, err := template.ParseGlob(globalConfig.TemplateDir + "*")
	if err != nil {
		return &Error{Code: ErrParseTemplates, innerError: err}
	}

	templateLock.Lock()
	defer templateLock.Unlock()
	allTemplates = newTemplate
	return nil
}
