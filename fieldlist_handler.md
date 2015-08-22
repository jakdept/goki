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

```nohighlight
<html>
	<head>
		<title>Search the Wiki</title>
		<link rel="stylesheet" type="text/css" href="/site/css/search.css">
	</head>
	<body>
		{{with .allFields}}
			{{range .}}
				<a href="./{{.}}">{{.}}</a>
			{{end}}
		{{else}}
		<!--
					Displaying {{ .Total }} hits - best score was {{.MaxScore}} search took {{.Took}} seconds to complete.
		-->
			<aside id='resultsArea'>
			{{range .Hits}}
				<div class='searchResult'>
					<div class='meter'>
						<span style="width:calc(100% * {{.Score}});"></span>
					</div>
					<a href="{{.ID}}" class='resultLink' target="pagePreview">
						<h3>
						{{.ID}}
						</h3>
					</a>
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

The data output to the template is one of two structures - depending on the request.

If the request did not feature a field value - it's a list of the values of the field - the dataset will be structured:

```go
struct {
	allFields []string
}
```

* `allFields` is an array of strings that simply contains each field

- - - - - - - - - - - - -

If the request did contain a field value - it's a list of all documents within that field - the dataset will be structured as:

```go
struct {
    Request  *SearchRequest
    Hits     search.DocumentMatchCollection
    Total    uint64
    MaxScore float64
    Took     time.Duration
    Facets   search.FacetResults
}
```

Where:

* `Request` is the request that was sent - likely not needed
* `Hits` is an array of hits - structure explained later
* `Total` is the number of results that were matched
* `MaxScore` is the highest result match score
* `Took` is the amount of time the search took
* `Facets` is not used

The Hits are each structured as:

```go
type DocumentMatchCollection []struct {
    ID        string
    Score     float64
    Expl      *Explanation
    Locations FieldTermLocationMap
    Fragments FieldFragmentMap
    Fields    map[string]interface{}
}
```

Of note in here:

* `Score` contains the score of the match - the higher the closer the match to the search query
* `Fields` is a map containing the fields:
	* `path` - `string` - URI path pointing to the page
	* `title` - `string` - the page title
	* `topic` - `string` - all of the topics for the page, seperated by spaces
	* `author` - `string` - all of the authors for the page, seperated by spaces
	* `modified` - `time.Time` - timestamp of the last modification for that page
