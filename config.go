package gnosis

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

var staticConfig *Config
var configLock = new(sync.RWMutex)

type GlobalSection struct {
	port     string
	hostname string
}

type ServerSection struct {
	path        string
	prefix      string
	defaultPage string
	serverType  string
	restricted  []string
}

type Config struct {
	global     GlobalSection
	mainserver ServerSection
	server     []ServerSection
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
		log.Println("Problem loading config file: ", err)
	}

	// UnMarshal the config file that was read in
	temp := new(Config)
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
