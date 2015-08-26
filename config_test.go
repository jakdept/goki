package gnosis

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
	// "bytes"
	// "io"
	"strings"
	// "net/http"
	"html/template"
	"net/http/httptest"
	"time"
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

	go func() {
		config = GetConfig()
	}()

	assert.Nil(t, config, "Config should current be blocked and shoult not have loaded.")
	configLock.Unlock()
	time.Sleep(10)
	assert.NotNil(t, config, "Config should have loaded - is no longer blocked.")
}

func TestRenderTemplate(t *testing.T) {
	var tests = []struct {
		input    string
		template string
		expected string
	}{
		{"abc", "debug.raw", "abc"},
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

		go func() {
			err = RenderTemplate(actualContent, testSet.template, testSet.input)
		}()

		if actualContent.Body.String() != "" {
			t.Errorf("response buffer should currently be empty, but it contains [%q]",
				actualContent.Body.String())
		} else {
			templateLock.Unlock()
			time.Sleep(10)

			if string(actualContent.Body.String()) != testSet.expected {
				t.Errorf("\nexpected [%q]\ngot      [%q]", testSet.expected,
					actualContent.Body.String())
			}
		}
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
	expectedError := errors.New("html/template: pattern matches no files: `./notadir/**`")
	if err != expectedError {
		t.Errorf("got the wrong error response\nexpecting [%q]\n      got [%q]",
			err, expectedError)
	}
}
