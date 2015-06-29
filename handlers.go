package gnosis

// package file will contain MarkdownHandler and RawHandler to handle incoming requests
// Whole thing needs to be written

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/blevesearch/bleve"
	// bleveHttp "github.com/blevesearch/bleve/http"
)

type Page struct {
	Title    string
	ToC      template.HTML
	Body     template.HTML
	Topics   template.HTML
	Keywords template.HTML
}

func MarkdownHandler(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error

	// break up the request parameters - for reference, regex is listed below
	//filteredRequest, err := wikiFilter.FindStringSubmatch(request.URL.Path)

	requestPath := strings.TrimPrefix(request.URL.Path, serverConfig.Prefix)

	// If the request is empty, set it to the default.
	//if requestPath == "" || requestPath == "/" {
	if requestPath == "" {
		requestPath = serverConfig.Default
	log.Printf("replaced the request path - Request path is [%s] of length [%d], comparing against [%s] of length [%d]", requestPath, len(requestPath), "", len(""))
	} else {
	log.Printf("did not replace - Request path is [%s] of length [%d], comparing against [%s] of length [%d]", requestPath, len(requestPath), "", len(""))
}

	// If the request doesn't end in .md, add that
	if !strings.HasSuffix(requestPath, ".md") {
		requestPath = requestPath + ".md"
	}

	pdata := new(PageMetadata)
	err = pdata.LoadPage(serverConfig.Path + requestPath)
	if err != nil {
		log.Printf("request [ %s ] points to an bad file target [ %s ] sent to server %s", request.URL.Path, requestPath, serverConfig.Prefix)
		http.Error(responsePipe, err.Error(), 404)
		return
	}

	if pdata.MatchedTag(serverConfig.Restricted) {
		log.Printf("request [ %s ] was against a page [ %s ] with a restricted tag", request.URL.Path, requestPath)
		http.Error(responsePipe, err.Error(), 403)
		return
	}

	// parse any markdown in the input
	body := template.HTML(bodyParseMarkdown(pdata.Page))
	toc := template.HTML(tocParseMarkdown(pdata.Page))
	keywords := pdata.PrintKeywords()
	topics := pdata.PrintTopics(serverConfig.TopicURL)

	// ##TODO## put this template right in the function call
	// Then remove the Page Struct above
	response := Page{Title: "", ToC: toc, Body: body, Keywords: keywords, Topics: topics}
	err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, response)
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
}

func FindExtension(s string) (string, error) {
	for i := len(s); i > 0; i-- {
		if string(s[i]) == "." {
			return s[i:], nil
		}
	}
	return "", errors.New("found no extension")
}

func RawHandler(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error

	request.URL.Path = strings.TrimPrefix(request.URL.Path, serverConfig.Prefix)

	// If the request is empty, set it to the default.
	if request.URL.Path == "" || request.URL.Path == "/" {
		request.URL.Path = serverConfig.Default
	}

	// If the request is a blocked restriction, shut it down.
	//extension, err := FindExtension(request.URL.Path)
	for _, restricted := range serverConfig.Restricted {
		if strings.HasSuffix(request.URL.Path, restricted) {
			log.Printf("request %s was improperly routed to the file handler with an disallowed extension %s", request.URL.Path, restricted)
			http.Error(responsePipe, "Request not allowed", 403)
			return
		}
	}

	// Load the file - 404 on failure.
	contents, err := ioutil.ReadFile(serverConfig.Path + request.URL.Path)
	if err != nil {
		log.Printf("request [ %s ] points to an bad file target sent to server %s - %v", request.URL.Path, serverConfig.Prefix, err)
		http.Error(responsePipe, err.Error(), 404)
		return
	}

	_, err = responsePipe.Write([]byte(contents))
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
	return
}

func SearchHandler(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error

	request.URL.Path = strings.TrimPrefix(request.URL.Path, serverConfig.Prefix)

	// if err = request.ParseForm(); err != nil {
	// 	log.Printf("error parsing search request, %v", err)
	// 	http.Error(responsePipe, err.Error(), 500)
	// 	return
	// }

	queryArgs := request.URL.Query()

	// debugging information
	for k, v := range queryArgs {
		log.Println("key:", k)
		log.Println("val:", strings.Join(v, ""))
	}

	query := bleve.NewQueryStringQuery(queryArgs["s"][0])
	searchRequest := bleve.NewSearchRequest(query)

	// validate the query
	err = searchRequest.Query.Validate()
	if err != nil {
		log.Printf("Error validating query: %v", err)
		http.Error(responsePipe, err.Error(), 400)
		return
	}

	log.Println("validated query")

	// log.Println(bleveHttp.IndexNames())

	return

	// index := bleveHttp.IndexByName(serverConfig.Default)
	index, err := bleve.Open(serverConfig.Path)
	defer index.Close()
	if index == nil {
		log.Printf("no such index '%s'", serverConfig.Default)
		http.Error(responsePipe, err.Error(), 404)
		return
	} else if err != nil {
		log.Printf("no such index '%s'", serverConfig.Path)
		http.Error(responsePipe, err.Error(), 404)
		log.Printf("problem opening index '%s' - %v", serverConfig.Path, err)
		return
	}

	log.Println("opened index")

	// execute the query
	searchResponse, err := index.Search(searchRequest)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		http.Error(responsePipe, err.Error(), 400)
		return
	}

	log.Println("ran query")

	err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, searchResponse)
	// err = RenderTemplate(responsePipe, serverConfig.Template, searchResponse)
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}

	log.Println("Responded")
}

func MakeHandler(handlerConfig ServerSection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch handlerConfig.ServerType {
		case "markdown":
			MarkdownHandler(w, r, handlerConfig)
		case "raw":
			RawHandler(w, r, handlerConfig)
		case "simpleSearch":
			SearchHandler(w, r, handlerConfig)
		default:
			log.Printf("Bad server type [%s]", handlerConfig.ServerType)
		}
	}
}
