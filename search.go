package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
)

// given a path to an index, and a name of field to check
// lists all unique values for that field in index
func ListField(indexPath, field string) ([]string, error) {
	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)

	facet := bleve.NewFacetRequest(field, 1)
	searchRequest.AddFacet("allValues", facet)
	if err := query.Validate(); err != nil {
		return nil, err
	}

	// Open the index
	index, err := bleve.Open(indexPath)
	defer index.Close()
	if index == nil {
		return nil, fmt.Errorf("no such index [%s]", indexPath)
	} else if err != nil {
		return nil, fmt.Errorf("problem opening index [%s] - %s", indexPath, err)
	}

	searchResults, err := index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	results := *new([]string)
	for _, oneTerm := range searchResults.Facets["allValues"].Terms {
		results = append(results, oneTerm.Term)
	}

	return results, nil
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

func CreateResponseData(rawResults bleve.SearchResult, pageOffset int, topics []string, authors []string) (SearchResponse, error) {
	var response SearchResponse

	response.TotalHits = int(rawResults.Total)
	response.PageOffset = pageOffset
	response.SearchTime = rawResults.Took
	response.Topics = topics
	response.Authors = authors
	for _, hit := range rawResults.Hits {
		var newHit SearchResponseResult

		newHit.Score = float64(hit.Score * 100 / rawResults.MaxScore)

		if _, isThere := hit.Fields["title"]; !isThere {
			newHit.Title = ""
		} else if str, ok := hit.Fields["title"].(string); ok {
			newHit.Title = str
		} else {
			return response, errors.New("returned title was not a string")
		}

		if _, isThere := hit.Fields["path"]; !isThere {
			newHit.URIPath = ""
		} else if str, ok := hit.Fields["path"].(string); ok {
			newHit.URIPath = str
		} else {
			return response, errors.New("returned path was not a string")
		}

		if _, isThere := hit.Fields["body"]; !isThere {
			newHit.Body = ""
		} else if str, ok := hit.Fields["body"].(string); ok {
			newHit.Body = str
		} else {
			return response, errors.New("returned body was not a string")
		}

		if _, isThere := hit.Fields["topic"]; !isThere {
			newHit.Topics = []string{}
		} else if str, ok := hit.Fields["topic"].(string); ok {
			newHit.Topics = strings.Split(str, " ")
		} else {
			return response, errors.New("returned topics were not a string")
		}

		if _, isThere := hit.Fields["keyword"]; !isThere {
			newHit.Keywords = []string{}
		} else if str, ok := hit.Fields["keyword"].(string); ok {
			newHit.Keywords = strings.Split(str, " ")
		} else {
			return response, errors.New("returned keywords were not a string")
		}

		if _, isThere := hit.Fields["author"]; !isThere {
			newHit.Authors = []string{}
		} else if str, ok := hit.Fields["author"].(string); ok {
			newHit.Authors = strings.Split(str, " ")
		} else {
			return response, errors.New("returned authors were not a string")
		}

		response.Results = append(response.Results, newHit)
	}

	return response, nil
}
