topic: handler
topic: index
topic: search
topic: config
keyword: category
Query Search Handler
====================

The query search handler was the easier of the two handlers to implement - it contains less configuration.

Query Syntax
------------

The Query Search Handler uses a aquery format different from other formats.

* By default, any text passed is searched across all fields.
 * `ca bundle` would search for `ca` and `bundle` in all documents.
* By using quotes, you can unify a phrase
 * `"ca bundle"` would search for `ca bundle` in that order in all documents.
* By prepending an indexed field like `topic`, you can search for that term within that field.
 * `ca bundle topic:ssl` would search with a speccial focus on the topic `ssl`.
* By default, anything in the term is optional. Required and Forbidden items can be added with `+` and `-`.
 * `ca bundle +topic:ssl -ssh` is going to search for `ca` and `bundle`, but each result must be of topic `ssl` and must not contain `ssh`.
* By using the `^` operator, you may elevate any item within the query.
 * `ca bundle^2` would search for `ca` and `bundle`, but any match to `bundle` would count 5x that of any match to `ca`.
* For numeric fields, `<` and `>` can be used to filter.

 [All of this is documented upstream here](http://www.blevesearch.com/docs/Query-String-Query/).


Configuration
-------------

```nohighlight
{
  "ServerType": "querysearch"
  "Prefix": "/search/",
  "Path": "/var/www/wiki.index",
  "Template": "search.html",
},
```

The elements can appear in any order, and like the rest of the config, this is JSON formatted.

* `ServerType` always `querysearch`
* `Prefix` the URL path to handle. The most specific Prefix path is used.
* `Path` - location of the index to list against
* `Template` - the template to build a response with

When the request is recieved, the search is validated.

The query string is passed to the handler n the request var `s`. Thus, a valid request for the above configuration might be:

`http://localhost/search/?s=searching`

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