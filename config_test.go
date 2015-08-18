package gnosis

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	// "time"
)

func TestSimpleConfig(t *testing.T) {

	defaultConfig := `{
  "Global": {
    "Port": "8080",
    "Hostname": "localhost"
  },
  "Server": [
	  {
      "Path": "/var/www/wiki/",
      "Prefix": "/",
      "Default": "index",
      "ServerType": "markdown",
      "Restricted": []
    }
  ]
}`

	filepath := path.Join(os.TempDir(), "simpleini.txt")
	f, err := os.Create(filepath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath)
	if _, err := f.WriteString(defaultConfig); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	success := LoadConfig(filepath)

	assert.True(t, success, "Default configuration should load without error.")

	config := GetConfig()

	//assert.Nil(t, config, "Config file could not be accessed")

	assert.Equal(t, config.Global.Port, "8080", "read Port value incorrectly")
	assert.Equal(t, config.Global.Hostname, "localhost", "read Hostname value incorrectly")

	assert.Equal(t, config.Server[0].Path, "/var/www/wiki/", "read Path value incorrectly")
	assert.Equal(t, config.Server[0].Prefix, "/", "read Prefix value incorrectly")
	assert.Equal(t, config.Server[0].Default, "index", "read Default page value incorrectly")
	assert.Equal(t, config.Server[0].ServerType, "markdown", "read ServerType value incorrectly")

	assert.Equal(t, len(config.Server[0].Restricted), 0, "incorrect number of restricted elements") // putting this comment here so sublime stops freaking out about a line with one character
}

/*
func TestGetConfig(t *testing.T) {
	// directly put a config in the spot
	staticConfig = &Config{
		Global:GlobalSection{
			Port: "8080",
			Hostname: "localhost",
			TemplateDir: "/templates/",
		},
		Redirects: []RedirectSection{
			RedirectSection{
				Requested: "source",
				Target: "dest",
				Code: 302,
			},
		},
		Server: []ServerSection{
			ServerSection{
				ServerType: "markdown",
				Prefix: "/",
				Path: "/var/www/",
				Default: "readme",
				Template: "wiki.html",
				Restricted: []string{},
			},
		},
		Indexes: IndexSection{
			WatchDirs: map[string]string{
				"/var/www/": "/",
			},
			WatchExtension: ".md",
			IndexPath: "/index/",
			IndexType: "en",
			IndexName: "wiki",
			Restricted: []string{},
		},
	}

	configLock.RWLoci()
	var config *Config
	go func{
	config = GetConfig()
	}()

	assert.Nil(t, config, "Config should current be blocked and shoult not have loaded.")
	configLock.RWUnlock()
	time.Sleep(10)
	assert.NotNil(t, config, "Config should have loaded - is no longer blocked.")
}
*/