package main

import (
	//"html/template"
	//"io/ioutil"
	"log"
	//"regexp"

	"net/http"

	"flag"

	//"github.com/JackKnifed/blackfriday"
	"github.com/JackKnifed/gnosis"
	//"github.com/russross/blackfriday"
)

var configFile = flag.String("config", "config.json", "specify a configuration file")

func main() {
	flag.Parse()

	// ##TODO## check for false returnear- if null, the config could not be loaded
	gnosis.LoadConfig(*configFile)

	config := gnosis.GetConfig()

	rawFiles := http.FileServer(http.Dir(config.Mainserver.Path))
	siteFiles := http.FileServer(http.Dir(config.Mainserver.Path))

	http.Handle("/raw/", http.StripPrefix("/raw/", rawFiles))
	http.Handle("/site/", http.StripPrefix("/site/", siteFiles))
	http.HandleFunc("/", gnosis.MarkdownHandler)

	log.Println(http.ListenAndServe(":"+config.Global.Port, nil))

}
