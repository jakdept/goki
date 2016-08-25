package main

import (
	"strings"
	"time"

	"github.com/blevesearch/bleve"
)

type SearchResponse struct {
	// AllFields []string
	TotalHits  int
	PageOffset int
	SearchTime time.Duration
	Topics     []string
	Authors    []string
	Results    []SearchResponseResult
}

type SearchResponseResult struct {
	Title    string
	URIPath  string
	Score    float64
	Topics   []string
	Keywords []string
	Authors  []string
	Body     string
}

func (i *Index) CreateResponseData(rawResults bleve.SearchResult, pageOffset int) (
	SearchResponse, error) {

	topics, err := i.ListField("topics")
	if err != nil {
		return SearchResponse{}, err
	}

	authors, err := i.ListField("authors")
	if err != nil {
		return SearchResponse{}, err
	}

	response := SearchResponse{
		TotalHits:  int(rawResults.Total),
		PageOffset: pageOffset,
		SearchTime: rawResults.Took,
		Topics:     topics,
		Authors:    authors,
	}

	for _, hit := range rawResults.Hits {
		var newHit SearchResponseResult

		newHit.Score = float64(hit.Score * 100 / rawResults.MaxScore)

		for _, field := range []string{"title", "path", "body"} {
			if _, isThere := hit.Fields[field]; isThere {
				if str, ok := hit.Fields["title"].(string); ok {
					switch field {
					case "title":
						newHit.Title = str
					case "path":
						newHit.Path = str
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

// lists all unique values for that field in index
func (i *Index) ListField(field string) ([]string, error) {
	searchRequest := bleve.NewSearchRequest(bleve.NewMatchAllQuery())
	facet := bleve.NewFacetRequest(field, 1)
	searchRequest.AddFacet("allValues", facet)

	searchResult, err := i.Query(searchRequest)
	if err != nil {
		return []string{}, &Error{Code: ErrListField, innerError: err}
	}

	results := *new([]string)
	for _, oneTerm := range searchResults.Facets["allValues"].Terms {
		results = append(results, oneTerm.Term)
	}
	return results, nil
}

func (i *Index) ListFieldValues(field, match string, pageSize, page int) (
	SearchResponse, error) {

	query := bleve.NewTermQuery(match).SetField(field)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = pageSize

	err := searchRequest.Query.Validate()
	if err != nil {
		return FieldValueList{}, &Error{Code: ErrInvalidQuery, innerError: err}
	}

	rawResult, err = i.Query(searchRequest)
	if err != nil {
		return FieldValueList{}, &Error{Code: ErrBadQuery, innerError: err}
	}

	result, err := CreateResponseData(*rawResult, page*pageSize, allTopics, allAuthors)
	if err != nil {
		return SearchResponse, &Error{Code: ErrFormatSearchResponse, innerError: err}
	}

	return result, nil

}

func (i *Index) FuzzySearch(term string, topics, authors []string,
	page, pageSize int) (SearchResponse, error) {

	var topicQuery, authorQuery []bleve.Query
	for _, eachTopic := range topics {
		topicQuery = append(topicQuery, bleve.NewTermQuery(eachTopic))
	}
	for _, eachAuthor := range authors {
		authorQuery = append(authorQuery, bleve.NewTermQuery(eachAuthor))
	}

	multiQuery := []bleve.Query{bleve.NewFuzzyQuery(term)}
	if len(topicQuery) > 0 {
		multiQuery = append(multiQuery, bleve.NewDisjunctionQuery(topicQuery))
	}
	if len(authorQuery) > 0 {
		multiQuery = append(multiQuery, bleve.NewDisjunctionQuery(authorQuery))
	}

	query := bleve.NewConjunctionQuery(multiQuery)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = pageSize
	searchRequest.From = pageSize * page

	err := searchRequest.Query.Validate()
	if err != nil {
		return SearchResponse{}, &Error{Code: ErrInvalidQuery, value: searchRequest.Query}
	}

	searchResult, err := i.Query(searchRequest)
	if err != nil {
		return SearchResponse{}, err
	}

	searchResponse, err := i.CreateResponseData(searchResult, page)
	if err != nil {
		return SearchResponse{}, err
	}

	return SearchResponse, nil
}

func (i *Index) QuerySearch(terms string, page, pageSize int) (SearchResponse, error) {
	query := bleve.NewQueryStringQuery(terms)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"path", "title", "topic", "author", "modified"}
	searchRequest.Size = pageSize
	searchRequest.From = pageSize * page

	err := searchRequest.Query.Validate()
	if err != nil {
		return SearchResponse{}, &Error{Code: ErrInvalidQuery, value: searchRequest.Query}
	}

	searchResult, err := i.Query(searchRequest)
	if err != nil {
		return SearchResponse{}, err
	}

	searchResponse, err := i.CreateResponseData(searchResult, page)
	if err != nil {
		return SearchResponse{}, err
	}

	return SearchResponse, nil
}
