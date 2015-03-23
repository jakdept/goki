package gnosis

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	//"strings"
	"testing"
)

func TestDefaultJsonConfig(t *testing.T) {

	defaultConfig := `{
  "Global": {
    "Port": "8080",
    "Hostname": "localhost"
  },
  "Mainserver": {
      "Path": "/var/www/wiki/",
      "Prefix": "/",
      "DefaultPage": "index",
      "ServerType": "markdown",
      "Restricted": [
      ]
    },
  "Server": [
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

	assert.Equal(t, success, "Default configuration should load without error.")

	config := GetConfig()

	//assert.Nil(t, config, "Config file could not be accessed")

	assert.Equal(t, config.Global.Port, "8080", "read Port value incorrectly")
	assert.Equal(t, config.Global.Hostname, "localhost", "read Hostname value incorrectly")

	assert.Equal(t, config.Mainserver.Path, "/var/www/wiki/", "read Path value incorrectly")
	assert.Equal(t, config.Mainserver.Prefix, "/", "read Prefix value incorrectly")
	assert.Equal(t, config.Mainserver.DefaultPage, "index", "read Default page value incorrectly")
	assert.Equal(t, config.Mainserver.ServerType, "markdown", "read ServerType value incorrectly")

	assert.Equal(t, len(config.Mainserver.Restricted), 0, "incorrect number of restricted elements") // putting this comment here so sublime stops freaking out about a line with one character
}

func TestSimpleJsonConfig(t *testing.T) {

	defaultConfig := `{
  "Global": {
    "Port": "8080",
    "Hostname": "wiki.hostbaitor.com"
  },
  "Mainserver": {
      "Path": "/var/www/wiki/",
      "Prefix": "/",
      "DefaultPage": "index",
      "ServerType": "markdown",
      "Restricted": [
        "internal",
        "handbook"
      ]
    },
  "Server": [
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

	assert.Equal(t, success, "Default configuration should load without error.")

	config := GetConfig()

	//assert.Nil(t, config, "Config file could not be accessed")

	assert.Equal(t, config.Global.Port, "8080", "read Port value incorrectly")
	assert.Equal(t, config.Global.Hostname, "localhost", "read Hostname value incorrectly")

	assert.Equal(t, config.Mainserver.Path, "/var/www/wiki/", "read Path value incorrectly")
	assert.Equal(t, config.Mainserver.Prefix, "/", "read Prefix value incorrectly")
	assert.Equal(t, config.Mainserver.DefaultPage, "index", "read Default page value incorrectly")
	assert.Equal(t, config.Mainserver.ServerType, "markdown", "read ServerType value incorrectly")

	assert.Equal(t, config.Mainserver.Restricted[0], "internal", "read first Restricted value incorrectly")
	assert.Equal(t, config.Mainserver.Restricted[1], "internal", "read first Restricted value incorrectly")

	assert.Equal(t, len(config.Mainserver.Restricted), 2, "incorrect number of restricted elements") // putting this comment here so sublime stops freaking out about a line with one character
}
