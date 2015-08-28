package gnosis

// package file will contain MarkdownHandler and RawHandler to handle incoming requests
// Whole thing needs to be written

import (
	// "errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	// bleveHttp "github.com/blevesearch/bleve/http"
)

type Page struct {
	Title    string
	ToC      template.HTML
	Body     template.HTML
	Topics   []string
	Keywords []string
	Authors  []string
}

func MarkdownHandler(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error

	requestPath := strings.TrimPrefix(request.URL.Path, serverConfig.Prefix)

	// If the request is empty, set it to the default.
	if requestPath == "" || requestPath == "/" {
		requestPath = serverConfig.Default
	}

	// If the request doesn't end in .md, add that
	if !strings.HasSuffix(requestPath, ".md") {
		requestPath = requestPath + ".md"
	}

	pdata := new(PageMetadata)
	err = pdata.LoadPage(serverConfig.Path + requestPath)
	if err != nil {
		log.Printf("request [ %s ] points to an bad file target [ %s ] sent to server %s",
			request.URL.Path, requestPath, serverConfig.Prefix)
		http.Error(responsePipe, "Page not Found", http.StatusNotFound)
		return
	}

	if pdata.MatchedTopic(serverConfig.Restricted) {
		log.Printf("request [ %s ] was against a page [ %s ] with a restricted tag",
			request.URL.Path, requestPath)
		http.Error(responsePipe, "Restricted Page", http.StatusNotFound)
		//http.Error(responsePipe, err.Error(), http.StatusForbidden)
		return
	}

	// parse any markdown in the input
	body := template.HTML(bodyParseMarkdown(pdata.Page))
	toc := template.HTML(tocParseMarkdown(pdata.Page))
	topics, keywords, authors := pdata.ListMeta()

	// ##TODO## put this template right in the function call
	// Then remove the Page Struct above
	response := Page{Title: pdata.Title, ToC: toc, Body: body, Keywords: keywords, Topics: topics, Authors: authors}
	err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, response)
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
}

func RawHandler(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error

	request.URL.Path = strings.TrimPrefix(request.URL.Path, serverConfig.Prefix)

	// If the request is empty, set it to the default.
	if request.URL.Path == "" || request.URL.Path == "/" {
		request.URL.Path = serverConfig.Default
	}

	// If the request is a blocked restriction, shut it down.
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
		log.Printf("request [ %s ] points to an bad file target sent to server %s - %v",
			request.URL.Path, serverConfig.Prefix, err)
		http.Error(responsePipe, err.Error(), 404)
		return
	}

	responsePipe.Header().Set("Content-Type", http.DetectContentType(contents))

	_, err = responsePipe.Write([]byte(contents))
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
	return
}

func SearchHandler(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error
	var ok bool

	request.URL.Path = strings.TrimPrefix(request.URL.Path, serverConfig.Prefix)

	queryArgs := request.URL.Query()

	if _, ok = queryArgs["s"]; !ok {
		err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, make([]bleve.SearchResult, 0))
		if err != nil {
			http.Error(responsePipe, err.Error(), 500)
		}
		return
	}

	page := 0
	if _, ok := queryArgs["page"]; ok {
		page, err = strconv.Atoi(queryArgs["page"][0])
		if err != nil {
			log.Printf("invalid page detected [%s] - %s", queryArgs["page"][0], err)
			page = 0
		}
	}

	pageSize := 50
	if _, ok := queryArgs["pagesize"]; ok {
		pageSize, err = strconv.Atoi(queryArgs["pagesize"][0])
		if err != nil {
			log.Printf("invalid pageSize detected [%s] - %s", queryArgs["pagesize"][0], err)
			pageSize = 0
		}
	}

	query := bleve.NewQueryStringQuery(queryArgs["s"][0])
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = pageSize
	searchRequest.From = pageSize * page

	// validate the query
	err = searchRequest.Query.Validate()
	if err != nil {
		log.Printf("Error validating query: %v", err)
		http.Error(responsePipe, err.Error(), 400)
		return
	}

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

	// execute the query
	searchResponse, err := index.Search(searchRequest)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		http.Error(responsePipe, err.Error(), 400)
		return
	}

	templateData, err := CreateResponseData(*searchResponse, page*pageSize)
	if err != nil {
		log.Printf("Error translating query results: %v", err)
		http.Error(responsePipe, err.Error(), 500)
	}

	err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, templateData)
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
}

func FuzzySearch(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error
	var ok bool

	request.URL.Path = strings.TrimPrefix(request.URL.Path, serverConfig.Prefix)

	queryArgs := request.URL.Query()

	if _, ok = queryArgs["s"]; !ok {
		err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, make([]bleve.SearchResult, 0))
		if err != nil {
			http.Error(responsePipe, err.Error(), 500)
		}
		return
	}

	// start with a string query
	var multiQuery []bleve.Query
	multiQuery[0] = bleve.NewFuzzyQuery(queryArgs["s"][0])

	// add in the required topics to the query
	var topics []bleve.Query
	for _, topic := range queryArgs["topic"] {
		topics = append(topics, bleve.NewTermQuery(topic))
	}
	if len(topics) > 0 {
		multiQuery = append(multiQuery, bleve.NewDisjunctionQuery(topics))
	}

	var authors []bleve.Query
	for _, author := range queryArgs["author"] {
		authors = append(authors, bleve.NewTermQuery(author))
	}
	if len(authors) > 0 {
		multiQuery = append(multiQuery, bleve.NewDisjunctionQuery(authors))
	}

	page := 0
	if _, ok := queryArgs["page"]; ok {
		page, err = strconv.Atoi(queryArgs["page"][0])
		if err != nil {
			log.Printf("invalid page detected [%s] - %s", queryArgs["page"][0], err)
			page = 0
		}
	}

	pageSize := 50
	if _, ok := queryArgs["pagesize"]; ok {
		pageSize, err = strconv.Atoi(queryArgs["pagesize"][0])
		if err != nil {
			log.Printf("invalid pageSize detected [%s] - %s", queryArgs["pagesize"][0], err)
			pageSize = 0
		}
	}

	query := bleve.NewConjunctionQuery(multiQuery)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = pageSize
	searchRequest.From = pageSize * page

	// validate the query
	err = searchRequest.Query.Validate()
	if err != nil {
		log.Printf("Error validating query: %v", err)
		http.Error(responsePipe, err.Error(), 400)
		return
	}

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

	// execute the query
	searchResponse, err := index.Search(searchRequest)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		http.Error(responsePipe, err.Error(), 400)
		return
	}

	templateData, err := CreateResponseData(*searchResponse, page*pageSize)
	if err != nil {
		log.Printf("Error translating query results: %v", err)
		http.Error(responsePipe, err.Error(), 500)
	}

	err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, templateData)
	if err != nil {
		http.Error(responsePipe, err.Error(), 500)
	}
}

func FieldListHandler(responsePipe http.ResponseWriter, request *http.Request, serverConfig ServerSection) {

	var err error
	fieldValue := ""

	urlWithoutPrefix := strings.TrimPrefix(request.URL.Path, serverConfig.Prefix)
	urlWithoutPrefix = strings.TrimPrefix(urlWithoutPrefix, "/")
	if len(urlWithoutPrefix) > 0 {
		temp := strings.SplitN(urlWithoutPrefix, "/", 2)
		if len(temp) > 0 {
			fieldValue = temp[0]
		}
	}

	if fieldValue == "" {
		// this is where I would put my facet listing thing
		// IF I HAD ONE
		// #TODO ^^ that ^^
		fields, err := ListField(serverConfig.Path, serverConfig.Default)
		if err != nil {
			http.Error(responsePipe, err.Error(), 500)
		}
		err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template,
			struct{ AllFields []string }{AllFields: fields})
		if err != nil {
			http.Error(responsePipe, err.Error(), 500)
		}
	} else {

		queryArgs := request.URL.Query()

		page := 0
		if _, ok := queryArgs["page"]; ok {
			page, err = strconv.Atoi(queryArgs["page"][0])
			if err != nil {
				log.Printf("invalid page detected [%s] - %s", queryArgs["page"][0], err)
				page = 0
			}
		}

		pageSize := 50
		if _, ok := queryArgs["pagesize"]; ok {
			pageSize, err = strconv.Atoi(queryArgs["pagesize"][0])
			if err != nil {
				log.Printf("invalid pageSize detected [%s] - %s", queryArgs["pagesize"][0], err)
				pageSize = 0
			}
		}

		query := bleve.NewTermQuery(fieldValue).SetField(serverConfig.Default)
		searchRequest := bleve.NewSearchRequest(query)

		searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
		searchRequest.Size = 1000

		// validate the query
		err = searchRequest.Query.Validate()
		if err != nil {
			log.Printf("Error validating query: %v", err)
			http.Error(responsePipe, err.Error(), 400)
			return
		}

		// Open the index
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

		// execute the query
		searchResponse, err := index.Search(searchRequest)
		if err != nil {
			log.Printf("Error executing query: %v", err)
			http.Error(responsePipe, err.Error(), 400)
			return
		}

		templateData, err := CreateResponseData(*searchResponse, page*pageSize)
		if err != nil {
			log.Printf("Error translating query results: %v", err)
			http.Error(responsePipe, err.Error(), 500)
		}

		err = allTemplates.ExecuteTemplate(responsePipe, serverConfig.Template, templateData)
		if err != nil {
			http.Error(responsePipe, err.Error(), 500)
		}
	}
}

func MakeHandler(handlerConfig ServerSection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToLower(handlerConfig.ServerType) {
		case "markdown":
			MarkdownHandler(w, r, handlerConfig)
		case "raw":
			RawHandler(w, r, handlerConfig)
		case "simplesearch":
			SearchHandler(w, r, handlerConfig)
		case "fieldlist":
			FieldListHandler(w, r, handlerConfig)
		case "fuzzysearch":
			FuzzySearch(w, r, handlerConfig)
		default:
			log.Printf("Bad server type [%s]", handlerConfig.ServerType)
		}
	}
}
