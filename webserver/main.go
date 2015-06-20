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

var indexes gnosisIndex[]

var configFile = flag.String("config", "config.json", "specify a configuration file")

var quitChan = make(chan os.Signal, 1)

func cleanup() {
	for _ = range quitChan {
		log.Println("Recieved an interrupt, shutting down")
		for index := indexes {
			index.closeIndex()
		}
	}
}

func main() {
	flag.Parse()

	// ##TODO## check for false returnear- if null, the config could not be loaded
	gnosis.LoadConfig(*configFile)

	config := gnosis.GetConfig()

	// set up my interrupt channel and go routine
	signal.Notify(quitChan, os.Interrupt)
	go cleanup()

	for _, individualIndex := range config.Indexes {
		index, err := gnosis.openIndex(individualIndex)
		if err != nil {
			log.Println(err)
		} else {
			indexes = append(indexes, index...)
		}
	}

	gnosis.ParseTemplates(config.Global)

	for _, redirect := range config.Redirects {
		http.Handle(redirect.Requested, http.RedirectHandler(redirect.Target, redirect.Code))
	}
	for _, individualServer := range config.Server {
		http.HandleFunc(individualServer.Prefix, gnosis.MakeHandler(individualServer))
	}

	log.Println(http.ListenAndServe(":"+config.Global.Port, nil))

}
