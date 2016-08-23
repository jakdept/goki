package main

import (
	"strings"
	"time"

	"github.com/blevesearch/bleve"
)

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

func (i *Index) FieldValueList(field, match string, pageSize, page int) (
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

	var response SearchResponse

	response.TotalHits = int(rawResults.Total)
	response.PageOffset = pageOffset
	response.SearchTime = rawResults.Took

	topics, err := i.ListField("topics")
	if err != nil {
		return SearchResponse{}, err
	}
	response.Topics = topics
	authors, err := i.ListField("authors")
	if err != nil {
		return SearchResponse{}, err
	}
	response.Authors = authors

	for _, hit := range rawResults.Hits {
		var newHit SearchResponseResult

		newHit.Score = float64(hit.Score * 100 / rawResults.MaxScore)

		if _, isThere := hit.Fields["title"]; !isThere {
			newHit.Title = ""
		} else if str, ok := hit.Fields["title"].(string); ok {
			newHit.Title = str
		} else {
			return SearchResponse{}, &Error{
				Code:  ErrResultsFormatType,
				path:  "title",
				value: hit.Fields["title"]}
		}

		if _, isThere := hit.Fields["path"]; !isThere {
			newHit.URIPath = ""
		} else if str, ok := hit.Fields["path"].(string); ok {
			newHit.URIPath = str
		} else {
			return SearchResponse{}, &Error{
				Code:  ErrResultsFormatType,
				path:  "path",
				value: hit.Fields["path"]}
		}

		if _, isThere := hit.Fields["body"]; !isThere {
			newHit.Body = ""
		} else if str, ok := hit.Fields["body"].(string); ok {
			newHit.Body = str
		} else {
			return SearchResponse{}, &Error{
				Code:  ErrResultsFormatType,
				path:  "body",
				value: hit.Fields["body"]}
		}

		if _, isThere := hit.Fields["topic"]; !isThere {
			newHit.Topics = []string{}
		} else if str, ok := hit.Fields["topic"].(string); ok {
			newHit.Topics = strings.Split(str, " ")
		} else {
			return SearchResponse{}, &Error{
				Code:  ErrResultsFormatType,
				path:  "topic",
				value: hit.Fields["topic"]}
		}

		if _, isThere := hit.Fields["keyword"]; !isThere {
			newHit.Keywords = []string{}
		} else if str, ok := hit.Fields["keyword"].(string); ok {
			newHit.Keywords = strings.Split(str, " ")
		} else {
			return SearchResponse{}, &Error{
				Code:  ErrResultsFormatType,
				path:  "keyword",
				value: hit.Fields["keyword"]}
		}

		if _, isThere := hit.Fields["author"]; !isThere {
			newHit.Authors = []string{}
		} else if str, ok := hit.Fields["author"].(string); ok {
			newHit.Authors = strings.Split(str, " ")
		} else {
			return SearchResponse{}, &Error{
				Code:  ErrResultsFormatType,
				path:  "author",
				value: hit.Fields["author"]}
		}

		response.Results = append(response.Results, newHit)
	}

	return response, nil
}
