topic:handler
topic: index
keyword: category
Field List Handler
================
The Field List handler is used to list all items with a specific field.

Configuration
-------------
An example config section for this handler:

```nohighlight
{
  "ServerType": "fieldList",
  "Prefix": "/topic/",
  "Path": "/var/www/wiki/",
  "Default": "readme",
  "Template": "wiki.html",
},
```

The elements can appear in any order, and like the rest of the config, this is JSON formatted.

* `ServerType` always `fieldList`
* `Prefix` the URL path to handle. The most specific Prefix path is used.
* `Path` - location of the index to list against
* `Default` - the field within that index to list - likely `topic` or `author`
* `Template` - the template to build a response with

When the request is recieved, it is validated. If it is valid, the Prefix is stripped off. If the next field is blank, all of the unique values of that field are listed; otherwise all entries from the index matching the first section are listed.

With the above configuration, `http://domain/topic/` would load a page listing all of the topics within the index, and `http://domain/topic/handler` would list all pages that have the `topic` of `handler`.

Example Template
----------------
An exmaple template for this handler is provided below:

```html
<html>
  <head>
    <title>Listing Topics</title>
    <link rel="stylesheet" type="text/css" href="/site/css/search.css">
  </head>
  <body>
    {{with .AllFields}}
    {{range .}}
    <a href="./{{.}}">
      <h2>{{.}}</h2>
    </a>
    {{end}}
    {{else}}
    <aside id='resultsArea'>
      {{ with .Results}}
      {{range .}}
      <div class='searchResult'>
        <a href="{{.URIPath}}" class='resultLink' target="pagePreview">
          <h2>
	          {{.Title}}
          </h2>
        </a>
      </div>
      {{else}}
      <h2>
      No Results found
      </h2>
      {{end}}
      {{else}}
      <aside id='resultsArea'>
          {{template "searchBox.html"}}
      </aside>
      {{end}}
    </aside>
    {{end}}
    <iframe id="pagePreview" name='pagePreview'></iframe>
  </body>
</html>
```

Go's documentation on [text templates](http://golang.org/pkg/text/template/) and [HTML templates](http://golang.org/pkg/html/template/) will likely be useful.

Output Data
-----------

The data output to the template is one of two structures - depending on the request.

If the request did not feature a field value - it's a list of the values of the field - the dataset will be structured:

```go
struct {
	AllFields []string
}
```

* `AllFields` is an array of strings that simply contains each field

- - - - - - - - - - - - -

If the request did contain a field value - it's a list of all documents within that field - the dataset will be structured as:

```go
type SearchResponse struct {
	AllFields []string
	TotalHits int
	MaxScore float64
	PageOffset int
	SearchTime time.Duration
	Results []SearchResponseResult
}
```

Where:

* `AllFields` is empty, allow differentiation
* `TotalHits` is the number of results that were matched
* `MaxScore` is the highest result match score
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
* `Score` contains the score of the match - the higher the closer the match to the search query
* `Topics` - all of the topics for the page
* `Keywords` - all of the keywords associated with the page
* `Author` - all of the authors for the page
* `Body` - contains a relevant portion of the page, with results highlighted
* `modified` - timestamp of the last modification for that page
