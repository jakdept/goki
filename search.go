package main

import (
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/blevesearch/bleve"
	blevequery "github.com/blevesearch/bleve/search/query"
)

// you probably want this for docs
// http://localhost:6060/pkg/github.com/JackKnifed/goki/vendor/github.com/blevesearch/bleve/#NewConjunctionQuery

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
	var err error

	switch match {
	case "":
		rawResult = &bleve.SearchResult{}
	default:
		query := bleve.NewTermQuery(match)
		query.SetField(field)
		searchRequest := bleve.NewSearchRequest(query)
		searchRequest.Fields = []string{
			"path",
			"title",
			"topic",
			"author",
			"modified",
		}
		searchRequest.Size = pageSize

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
	Term     string   `form:"s,omitempty"`
	Topics   []string `form:"topic,omitempty"`
	Authors  []string `form:"author,omitempty"`
	Page     int      `form:"page,omitempty"`
	PageSize int      `form:"pageSize,omitempty"`
}

// FuzzySearch runs a fuzzy search with the given input parameters against
//  the given query
func FuzzySearch(i Index, v FuzzySearchValues) (SearchResponse, error) {
	var disjunctionQuery, conjunctionQuery []blevequery.Query
	for _, eachTopic := range v.Topics {
		newQuery := bleve.NewTermQuery(eachTopic)
		newQuery.SetField("topic")
		disjunctionQuery = append(disjunctionQuery, newQuery)
	}
	for _, eachAuthor := range v.Authors {
		newQuery := bleve.NewTermQuery(eachAuthor)
		newQuery.SetField("author")
		disjunctionQuery = append(disjunctionQuery, newQuery)
	}

	if v.Term != "" {
		conjunctionQuery = append(conjunctionQuery, bleve.NewFuzzyQuery(v.Term))
	}

	switch {
	case len(disjunctionQuery) > 1:
		conjunctionQuery = append(conjunctionQuery, bleve.NewDisjunctionQuery(disjunctionQuery...))
		conjunctionQuery[len(conjunctionQuery)-1].(*blevequery.DisjunctionQuery).SetMin(1)
	case len(disjunctionQuery) == 1:
		conjunctionQuery = append(conjunctionQuery, disjunctionQuery[0])
	}

	var query blevequery.Query
	switch {
	case len(conjunctionQuery) == 0:
		query = bleve.NewMatchAllQuery()
	case len(conjunctionQuery) == 1:
		query = conjunctionQuery[0]
	default:
		query = bleve.NewConjunctionQuery(conjunctionQuery...)
	}

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{
		"path",
		"title",
		"topic",
		"author",
		"modified",
		"body",
	}
	searchRequest.Highlight = bleve.NewHighlight()
	searchRequest.Highlight.AddField("body")

	rawResult, err := i.Query(searchRequest)
	if err != nil {
		return SearchResponse{}, &Error{Code: ErrInvalidQuery, innerError: err}
	}

	for id, each := range rawResult.Hits {
		oldBody, ok := each.Fields["body"].(string)
		if !ok {
			continue
		}
		if len(oldBody) > 480 {
			newBody := oldBody[:480]
			for !unicode.IsSpace(rune(newBody[len(newBody)-1])) {
				if len(newBody) < 100 {
					break
				}
				newBody = newBody[:len(newBody)-2]
			}
			newBody += "..."
			rawResult.Hits[id].Fields["body"] = newBody
		}
	}

	searchResult, err := CreateResponseData(i, rawResult, v.Page)
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
