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

	"github.com/gorilla/handlers"
	//"github.com/JackKnifed/blackfriday"
	//"github.com/JackKnifed/goki"
	//"github.com/russross/blackfriday"
)

var configFile = flag.String("config", "config.json", "specify a configuration file")

var quitChan = make(chan os.Signal, 1)

// func cleanup() {
// 	_ = <-quitChan
// 	log.Println("Recieved an interrupt, shutting down")
// 	for _, index := range indexes {
// 		index.CloseIndex()
// 	}
// 	os.Exit(0)
// }

func main() {
	flag.Parse()

	// ##TODO## check for false returnear- if null, the config could not be loaded
	if err := LoadConfig(*configFile); err != nil {
		log.Fatal(err)
	}

	config := GetConfig()
	closer := make(chan struct{})

	// set up my interrupt channel and go routine
	// signal.Notify(quitChan, os.Interrupt)
	// go func() {
	// 	<-closer
	// 	<-quitChan
	// 	close(closer)
	// }()

	if err := ParseTemplates(*config); err != nil {
		log.Fatalf("Error parsing templates, %s", err)
	}

	mux, err := BuildMuxer(*config, closer, log.New(os.Stdout, "", 0))
	if err != nil {
		log.Fatal(err)
	}

	mux = handlers.LoggingHandler(os.Stdout, mux)
	canonical := handlers.CanonicalHost(config.Hostname, http.StatusMovedPermanently)

	log.Println(http.ListenAndServe(config.Address+":"+config.Port, canonical(mux)))
}
