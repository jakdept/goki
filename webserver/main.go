package main

import (
	//"html/template"
	//"io/ioutil"
	"log"
	//"regexp"
	"os"

	// "os/signal"
	"net/http"

	"flag"

	//"github.com/JackKnifed/blackfriday"
	"github.com/JackKnifed/goki"
	//"github.com/russross/blackfriday"
)

var configFile = flag.String("config", "config.json", "specify a configuration file")

var quitChan = make(chan os.Signal, 1)

/*
 func cleanup() {
	_ = <- quitChan
		log.Println("Recieved an interrupt, shutting down")
		for _, index := range indexes {
			index.CloseIndex()
		}
}
*/

func main() {
	flag.Parse()

	// ##TODO## check for false returnear- if null, the config could not be loaded
	if success := goki.LoadConfig(*configFile); success == false {
		log.Fatal("Could not parse the config, abandoning")
	}

	config := goki.GetConfig()

	// set up my interrupt channel and go routine
	// signal.Notify(quitChan, os.Interrupt)
	// go cleanup()

	for _, individualIndex := range config.Indexes {
		// #TODO change this?
		if !goki.EnableIndex(individualIndex) {
			log.Fatalf("Failed opening index %s, abandoning", individualIndex.IndexPath)
		}
	}

	if err := goki.ParseTemplates(config.Global); err != nil {
		log.Fatalf("Error parsing templates, %s", err)
	}

	for _, redirect := range config.Redirects {
		http.Handle(redirect.Requested, http.RedirectHandler(redirect.Target, redirect.Code))
	}
	for _, individualServer := range config.Server {
		http.HandleFunc(individualServer.Prefix, goki.MakeHandler(individualServer))
	}

	log.Println(http.ListenAndServe(":"+config.Global.Port, nil))

}
