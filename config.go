package gnosis

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"sync"
)

var staticConfig *Config
var configLock = new(sync.RWMutex)

//var templates = template.Must(template.ParseFiles("/var/wiki-backend/wiki.html"))
//template.New("allTemplates")
var allTemplates = template.New("allTemplates")
var templateLock = new(sync.RWMutex)

type GlobalSection struct {
	Port        string
	Hostname    string
	TemplateDir string
	Templates   []string
	Redirects   map[string]string
}

type ServerSection struct {
	Path        string
	Prefix      string
	DefaultPage string
	Template    string
	ServerType  string
	Restricted  []string
}

type Config struct {
	Global     GlobalSection
	Mainserver ServerSection
	Server     []ServerSection
}

var defaultConfig = []byte(`{
  "Global": {
    "Port": "8080",
    "Hostname": "localhost"
  },
  "Mainserver": {
      "Path": "/var/www/wiki/",
      "Prefix": "/",
      "DefaultPage": "index",
      "ServerType": "markdown",
      "Template": "wiki.html",
      "Restricted": [
        "internal",
        "handbook"
      ]
    },
  "Server": [
  ]
}`)

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

	err = json.Unmarshal(defaultConfig, temp)

	if err != nil {
		log.Println("problem parsing built in default configuration - this should not happen")
		return false
	}

	err = json.Unmarshal(fileContents, temp)
	//Make sure you were able to read it in
	if err != nil {
		log.Printf("parse config error: %s", err.Error())
		return false
	}

	configLock.Lock()
	staticConfig = temp
	configLock.Unlock()

	return true
}

func ParseTemplates(globalConfig GlobalSection) {
	var err error
	newTemplate := template.New("newTemplate")

	for _, templateFile := range globalConfig.Templates {
		nextTemplate, err := newTemplate.ParseFiles(globalConfig.TemplateDir + templateFile)
		if nextTemplate != nil {
			newTemplate = nextTemplate
		} else {
			log.Printf("Found an invalid template, abandoning updating templates")
		}
	}

	if err == nil {
		templateLock.Lock()
		defer templateLock.Unlock()
		allTemplates = newTemplate
	}

}
