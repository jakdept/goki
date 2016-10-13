topic: handler
topic: index
topic: search
topic: config
keyword: category
Query Search Handler
====================

The fuzzy search handler is the more traditional search handler. It allows you to work with individual items.

Query URL Syntax
------------

THe fuzzy search handler works a bit differently. An example URL to hit it is:

```
http://localhost/search/?s=query&topic=apache&author=person
```

* The `topic` field can appear multiple times - specifying multiple topics.
 * At least one `topic` for the page must match one `topic` provided in the query.
 * Any articles that do not match this condition are excluded from the results.
* The `author` field can appear multiple times - specifying multiple authors.
 * At least one `author` for the page must match one `author` provided in the query.
 * Any articles that do not match this condition are excluded from the results.
* After this, the `s` field is individual term searches against fields - with decreasing priority.
 * `title` has the highest priority - a match in `title` gives the match the strongest score.
 * `keyword` has the next highest priority - a match in `keyword` helps quite a bit.
 * `URIPath` has the next highest priority.
 * `topic` has the next highest priority.
 * The `body` of the page has the next highest priority.
 * Finally, `author` has the lowest priority for the `s` search.

Configuration
-------------

```nohighlight
{
  "ServerType": "fuzzySearch"
  "Prefix": "/search/",
  "Template": "search.html",
	"FallbackTemplate": search.html"
}
```

The elements can appear in any order, and like the rest of the config, this is JSON formatted.

* `ServerType` always `fuzzySearch`
* `Prefix` the URL path to handle. The most specific Prefix path is used.
* `Template` - the template to build a response with
* `FallbackTemplate` - the template used to build a response if no search or no results

When the request is recieved, the search is validated.

The query string is passed to the handler in the var `s`.
Thus, a valid request for the above configuration might be:

```
http://localhost/search/?s=searching
```

Example Template
----------------
An exmaple template for this handler is provided below:

```
<html>
	<head>
		<title>Search the Wiki</title>
		<link rel="stylesheet" type="text/css" href="/site/css/search.css">
		<link rel="stylesheet" type="text/css" href="/site/css/simplex.css">
	</head>
	<body>
		{{with .}}
		Displaying {{len .Results}} hits of {{ .TotalHits }} hits, starting with {{.PageOffset}} - search took {{.SearchTime}} seconds to complete.
			<aside id='resultsArea'>
		{{template "searchBox.html"}}
			{{range .Results}}
				<div class='searchResult'>
					<div class="progress">
					  <div class="progress-bar progress-bar-success" style="width: {{.Score}}%">
					  </div>
					</div>
					<a href="{{.URIPath}}" class='resultLink' target="pagePreview">
						<h3>
						{{.Title}}
						</h3>
					</a>
					<div class='well well-sm'>
						<div class='pageLocation'>
							<a href="{{.URIPath}}">{{.URIPath}}</a>
						</div>
						{{range .Topics}}
						<span class="label label-info">
							{{.}}
						</span>
						{{end}}
						{{range .Authors}}
						<a class='btn btn-sm btn-success' href="/author/{{.}}">
							{{.}}
						</a>
						{{end}}
					</div>
				</div>
			{{else}}
				<h3>
				No Results found
				</h3>
			{{end}}
		{{else}}
			<aside id='resultsArea'>
			{{template "searchBox.html"}}
			</aside>
		{{end}}
		</aside>
		<iframe id="pagePreview" name='pagePreview'></iframe>
	</body>
</html>
```

Go's documentation on [text templates](http://golang.org/pkg/text/template/) and [HTML templates](http://golang.org/pkg/html/template/) will likely be useful.

Output Data
-----------

The output to the template will be the search results:

```go
type SearchResponse struct {
	TotalHits int
	PageOffset int
	SearchTime time.Duration
	Results []SearchResponseResult
}
```

Where:

* `TotalHits` is the number of results that were matched
* `PageOffset` is the number of results skipped before the current page
* `SearchTime` is the amount of time the search took
* `Results` is an array of hits - structure explained later

The Hits are each structured as:

```go
type SearchResponseResult struct {
	Title string
	URIPath string
	Score float64
	Topics []string
	Keywords []string
	Authors []string
	Body string
}
```

Of note in here:

* `Title` - the page title
* `URIPath` - URI path pointing to the page
* `Score` contains the score of the match - a percentage
* `Topics` - all of the topics for the page
* `Keywords` - all of the keywords associated with the page
* `Author` - all of the authors for the page
* `Body` - contains a relevant portion of the page, with results highlighted
* `modified` - timestamp of the last modification for that page