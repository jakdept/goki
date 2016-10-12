package main

import (
	"bleve"
	"log"
	"net/http"
	"strings"
	"time"
)

// SearchResponse is the parent type structure that will come back to all
//  requests. []Results will contain child results.
type SearchResponse struct {
	TotalHits  int
	PageOffset int
	SearchTime time.Duration
	Topics     []string
	Authors    []string
	Results    []SearchResponseResult
}

// SearchResponseResult is the child type that will come back to all responses
type SearchResponseResult struct {
	Title    string
	URIPath  string
	Score    float64
	Topics   []string
	Keywords []string
	Authors  []string
	Body     string
}

// CreateResponseData takes a search result, and produces a SearchResponse
//  suitable for passing to a template.
func CreateResponseData(i Index, results *bleve.SearchResult, pageOffset int) (
	SearchResponse, error) {

	topics, err := ListField(i, "topics")
	if err != nil {
		return SearchResponse{}, err
	}

	authors, err := ListField(i, "authors")
	if err != nil {
		return SearchResponse{}, err
	}

	response := SearchResponse{
		TotalHits:  int(results.Total),
		PageOffset: pageOffset,
		SearchTime: results.Took,
		Topics:     topics,
		Authors:    authors,
	}

	for _, hit := range results.Hits {

		log.Printf("working with \n %#v\n", hit)

		var newHit SearchResponseResult

		newHit.Score = float64(hit.Score * 100 / results.MaxScore)

		for _, field := range []string{
			"title",
			"path",
			"body",
			"topic",
			"keyword",
			"author",
		} {
			if _, isThere := hit.Fields[field]; isThere {
				if str, ok := hit.Fields[field].(string); ok {
					switch field {
					case "title":
						newHit.Title = str
					case "path":
						newHit.URIPath = str
					case "body":
						newHit.Body = str
					case "topic":
						newHit.Topics = strings.Split(str, " ")
					case "keyword":
						newHit.Keywords = strings.Split(str, " ")
					case "author":
						newHit.Authors = strings.Split(str, " ")
					}
				} else {
					return SearchResponse{}, &Error{
						Code:  ErrResultsFormatType,
						path:  field,
						value: hit.Fields[field]}
				}
			}
		}

		log.Printf("working with \n %#v\n", newHit)
		response.Results = append(response.Results, newHit)
	}
	return response, nil
}

// ListField lists all unique values for that field in index
func ListField(i Index, field string) ([]string, error) {
	searchRequest := bleve.NewSearchRequest(bleve.NewMatchAllQuery())
	facet := bleve.NewFacetRequest(field, 1)
	searchRequest.AddFacet("allValues", facet)

	searchResult, err := i.Query(searchRequest)
	if err != nil {
		return []string{}, &Error{Code: ErrListField, innerError: err}
	}

	results := *new([]string)
	for _, oneTerm := range searchResult.Facets["allValues"].Terms {
		results = append(results, oneTerm.Term)
	}
	return results, nil
}

func ListAllField(i Index, field, match string, pageSize, page int) (
	SearchResponse, error) {

	var rawResult *bleve.SearchResult

	switch match {
	case "":
		rawResult = &bleve.SearchResult{}
	default:
		query := bleve.NewTermQuery(match).SetField(field)
		searchRequest := bleve.NewSearchRequest(query)
		searchRequest.Fields = []string{
			"path",
			"title",
			"topic",
			"author",
			"modified",
		}
		searchRequest.Size = pageSize

		err := searchRequest.Query.Validate()
		if err != nil {
			return SearchResponse{}, &Error{Code: ErrInvalidQuery, innerError: err}
		}

		rawResult, err = i.Query(searchRequest)
		if err != nil {
			return SearchResponse{}, &Error{Code: ErrInvalidQuery, innerError: err}
		}
	}

	result, err := CreateResponseData(i, rawResult, page*pageSize)
	if err != nil {
		return SearchResponse{}, &Error{
			Code:       ErrFormatSearchResponse,
			innerError: err,
		}
	}

	return result, nil
}

//FuzzySearchValues gives a standard structure to decode and pass to FuzzySearch
type FuzzySearchValues struct {
	s        string   `form:"s"`
	topics   []string `form:"topic"`
	authors  []string `form:"author"`
	page     int      `form:"page"`
	pageSize int      `form:"pageSize"`
}

// FuzzySearch runs a fuzzy search with the given input parameters against
//  the given query
func FuzzySearch(i Index, v FuzzySearchValues) (SearchResponse, error) {
	var topicQuery, authorQuery []bleve.Query
	for _, eachTopic := range v.topics {
		topicQuery = append(topicQuery, bleve.NewTermQuery(eachTopic))
	}
	for _, eachAuthor := range v.authors {
		authorQuery = append(authorQuery, bleve.NewTermQuery(eachAuthor))
	}

	multiQuery := []bleve.Query{bleve.NewFuzzyQuery(v.s)}
	if len(topicQuery) > 0 {
		multiQuery = append(multiQuery, bleve.NewDisjunctionQuery(topicQuery))
	}
	if len(authorQuery) > 0 {
		multiQuery = append(multiQuery, bleve.NewDisjunctionQuery(authorQuery))
	}

	query := bleve.NewConjunctionQuery(multiQuery)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = v.pageSize
	searchRequest.From = v.pageSize * v.page

	err := searchRequest.Query.Validate()
	if err != nil {
		return SearchResponse{}, &Error{
			Code:  ErrInvalidQuery,
			value: searchRequest.Query,
		}
	}

	rawResult, err := i.Query(searchRequest)
	if err != nil {
		return SearchResponse{}, err
	}

	searchResult, err := CreateResponseData(i, rawResult, v.page)
	if err != nil {
		return SearchResponse{}, err
	}

	return searchResult, nil
}

// QuerySearch runs a given query search and returns a SearchResponse against
//  the given index
func QuerySearch(i Index, terms string, page, pageSize int) (
	SearchResponse, error) {
	query := bleve.NewQueryStringQuery(terms)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = pageSize
	searchRequest.From = pageSize * page

	err := searchRequest.Query.Validate()
	if err != nil {
		return SearchResponse{}, &Error{
			Code:  ErrInvalidQuery,
			value: searchRequest.Query,
		}
	}

	rawResult, err := i.Query(searchRequest)
	if err != nil {
		return SearchResponse{}, err
	}

	searchResult, err := CreateResponseData(i, rawResult, page)
	if err != nil {
		return SearchResponse{}, err
	}

	return searchResult, nil
}

// FallbackSearchResponse is a function that writes a "bailout" template
func FallbackSearchResponse(i Index, w http.ResponseWriter,
	template string) {
	authors, err := ListField(i, "author")
	if err != nil {
		http.Error(w, "failed to list authors", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	topics, err := ListField(i, "topic")
	if err != nil {
		http.Error(w, "failed to list topics", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	fields := SearchResponse{Topics: topics, Authors: authors}

	err = allTemplates.ExecuteTemplate(w, template, fields)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	return
}
