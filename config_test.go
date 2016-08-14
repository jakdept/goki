package goki

import (
	"html/template"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.NoError(t, success, "Default configuration should load without error.")

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

func TestGetConfig(t *testing.T) {
	// directly put a config in the spot
	staticConfig = &Config{
		Global: GlobalSection{
			Port:        "8080",
			Hostname:    "localhost",
			TemplateDir: "/templates/",
		},
		Redirects: []RedirectSection{
			RedirectSection{
				Requested: "source",
				Target:    "dest",
				Code:      302,
			},
		},
		Server: []ServerSection{
			ServerSection{
				ServerType: "markdown",
				Prefix:     "/",
				Path:       "/var/www/",
				Default:    "readme",
				Template:   "wiki.html",
				Restricted: []string{},
			},
		},
		Indexes: []IndexSection{
			IndexSection{
				WatchDirs: map[string]string{
					"/var/www/": "/",
				},
				WatchExtension: ".md",
				IndexPath:      "/index/",
				IndexType:      "en",
				IndexName:      "wiki",
				Restricted:     []string{},
			},
		},
	}

	configLock.Lock()

	var config *Config

	ch := make(chan *Config)

	go func() {
		ch <- GetConfig()
	}()

	select {
	case config = <-ch:
	default:
	}

	assert.Nil(t, config, "Config should current be blocked and shoult not have loaded.")
	configLock.Unlock()
	select {
	case config = <-ch:
	}
	assert.NotNil(t, config, "Config should have loaded - is no longer blocked.")
}

func TestRenderTemplate(t *testing.T) {
	var tests = []struct {
		input    string
		template string
		expected string
	}{
		{"abc", "debug.raw", "abc"},
		{"other stuff", "debug.raw", "other stuff"},
	}

	var err error
	allTemplates, err = template.ParseGlob("./templates/*")
	if err != nil {
		t.Errorf("failed to parse templates - %s", err)
	}

	for _, testSet := range tests {
		templateLock.Lock()

		actualContent := httptest.NewRecorder()

		if allTemplates.Lookup(testSet.template) == nil {
			t.Errorf("did not find template [%q] that is needed", testSet.template)
		}

		ch := make(chan string)

		go func() {
			err = RenderTemplate(actualContent, testSet.template, testSet.input)
			ch <- actualContent.Body.String()
		}()

		var testResponse string

		select {
		case testResponse = <-ch:
		default:
		}

		assert.Empty(t, testResponse, "expecte empty testResponse, has [%q]", testResponse)

		templateLock.Unlock()
		select {
		case testResponse = <-ch:
		}

		assert.Equal(t, testSet.expected, testResponse,
			"expecte empty testResponse, has [%q]", testResponse)

	}
}

func TestParseTemplates(t *testing.T) {
	ParseTemplates(GlobalSection{TemplateDir: "./templates/*"})

	files, err := filepath.Glob("./templates/*")
	if err != nil {
		t.Error(err)
	}
	if len(files) != len(allTemplates.Templates()) {
		t.Error("There are unparsed templates in the direcory")
	}

	for _, templateName := range files {
		if allTemplates.Lookup(strings.TrimPrefix(templateName, "templates/")) == nil {
			t.Errorf("Could not find template %s", templateName)
		}
	}
}

func TestBadParseTemplates(t *testing.T) {
	err := ParseTemplates(GlobalSection{TemplateDir: "./notadir/*"})
	localError, ok := err.(*Error)
	assert.True(t, ok, "did not get back my type of error")
	assert.Equal(t, localError.Code, ErrParseTemplates, "got the wrong error response - %#v")
}
