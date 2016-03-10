package goki

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var staticConfig *Config
var configLock = new(sync.RWMutex)

type Config struct {
	Global    GlobalSection
	Redirects []RedirectSection
	Server    []ServerSection
	Indexes   []IndexSection
}

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
}

type RedirectSection struct {
	Requested string
	Target    string
	Code      int
}

type IndexSection struct {
	WatchDirs      map[string]string // Location that we will be watching for updates
	WatchExtension string            // file extensions that we will watch for within that dir
	IndexPath      string            //location to store the index
	IndexType      string            // type of index - likely "en"
	IndexName      string            // name of the index
	Restricted     []string          // Tags to restrict indexing on
}

type ServerSection struct {
	Path       string   // filesystem path to serve out
	Prefix     string   // Web URL Prefix - alternatively the prefix for a search handler
	Default    string   // Default page to serve if empty URI - alternatively the facet to list against
	Template   string   // Template file to build the response from
	ServerType string   // markdown, raw, search, or facet to denote the type of Server handle
	TopicURL   string   // URI prefix to redirect to topic pages
	Restricted []string // list of restricts - extensions for raw, topics for markdown
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

	CleanConfig(temp)

	log.Println(temp.Indexes)

	configLock.Lock()
	staticConfig = temp
	configLock.Unlock()

	return true
}

func CleanConfig(config *Config) {
	if !strings.HasSuffix(config.Global.TemplateDir, string(os.PathSeparator)) {
		config.Global.TemplateDir = config.Global.TemplateDir + string(os.PathSeparator)
	}
	for _, indexSection := range config.Indexes {
		for origDirPath, origWebPath := range indexSection.WatchDirs {
			newDirPath := strings.TrimSuffix(origDirPath, string(os.PathSeparator))
			newWebPath := strings.TrimSuffix(origWebPath, "/")
			if newDirPath == "" {
				newDirPath = string(os.PathSeparator)
			}
			if newWebPath == "" {
				newWebPath = "/"
			}
			if newDirPath != origDirPath || newWebPath != origWebPath {
				delete(indexSection.WatchDirs, origDirPath)
				indexSection.WatchDirs[newDirPath] = newWebPath
			}
		}
	}
	for _, serverSection := range config.Server {
		if serverSection.Path != string(os.PathSeparator) {
			serverSection.Path = strings.TrimSuffix(serverSection.Path, string(os.PathSeparator))
		}
		if serverSection.Prefix != "/" {
			serverSection.Prefix = strings.TrimSuffix(serverSection.Prefix, "/")
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
	log.Printf("Parsing templates in [%q]", globalConfig.TemplateDir)
	newTemplate, err := template.ParseGlob(globalConfig.TemplateDir + "*")
	if err != nil {
		return err
	}

	loadedTemplates := newTemplate.Templates()
	for _, individualTemplate := range loadedTemplates {
		log.Printf("Loaded template %s ", individualTemplate.Name())
	}

	templateLock.Lock()
	defer templateLock.Unlock()
	allTemplates = newTemplate
	return nil
}
