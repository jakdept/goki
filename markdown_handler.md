topic:handler
Markdown Handler
================
The Markdown handler is the primary handler of this wiki.

Configuration
-------------
An example config section for this handler:

```nohighlight
{
  "ServerType": "markdown",
  "Prefix": "/",
  "Path": "/var/www/wiki/",
  "Default": "readme",
  "Extension": ".md"
  "Restricted": [
    "internal",
    "handbook"
  ]
  "Template": "wiki.html",
},
```

The elements can appear in any order, and like the rest of the config, this is JSON formatted.

* `ServerType` is always `markdown`
* `Prefix` is the URL path to handle. The most specific Prefix path is used.
* `Path` - the phyiscal path to the files on the system.
* `Default` - the default page to open (for a request that matches the Prefix)
* `Extension` - the file extension on the end of files
* `Restricted` - an array of topics that cannot appear 
* `Template` - the template to build a response from

When the request is recieved, it is validated. If it is valid, the Prefix is stripped off, the Path is added to the front, and if needed, the Default and Extension are loaded.

With the above configuration, `http://domain/` would load the page that would also reside at `http://domain/readme.md`.

Example Template
----------------
An exmaple template for this handler is provided below:

```nohighlight
<html>
	<head>
		<title>{{.Title}}</title>
		<link rel="stylesheet" type="text/css" href="/site/css/simplex.css">
		<link rel="stylesheet" type="text/css" href="/site/css/gnosis.css">
		<link rel="stylesheet" type="text/css" href="/site/css/highlight/default.css">
		<script src="/site/js/highlight.pack.js"></script>
		<script>hljs.initHighlightingOnLoad();</script>
		{{range .Keywords}}
			<meta name="keywords" content="{{.}}">
		{{end}}
	</head>
	<body>
		<aside id="sidebar">
			<!-- {{template "searchBox.html"}} -->
			<ul class="nav nav-pills nav-stacked">
			{{.ToC}}
			</ul>
			<div class="panel panel-info">
			  <div class="panel-heading">
				  <a href="/topic/">
				    <h3 class="panel-title">Topics on this page</h3>
			    </a>
			  </div>
			  <div class="panel-body">
					{{range .Topics}}
						<a class="btn btn-sm btn-info" href="/topic/{{.}}/">
							{{.}}
						</a>
					{{end}}
			  </div>
			</div>
			<div class="panel panel-success">
			  <div class="panel-heading">
				  <a href="/author/">
				    <h3 class="panel-title">This page written by</h3>
			    </a>
			  </div>
			  <div class="panel-body">
					{{range .Topics}}
						<a class="btn btn-sm btn-success" href="/author/{{.}}/">
							{{.}}
						</a>
					{{end}}
			  </div>
			</div>
		</aside>
		<div id="body" class="markdown-body">
		<h1>{{.Title}}</h1>
		{{.Body}}
		</div>
	</body>
</html>
```

Go's documentation on [text templates](http://golang.org/pkg/text/template/) and [HTML templates](http://golang.org/pkg/html/template/) will likely be useful.

Output Data
-----------

The following object is passed to that template:

```go
type Page struct {
	Title    string
	ToC      template.HTML
	Body     template.HTML
	Topics   []string
	Keywords []string
	Authors  []string
}
```

* `Title` is the page title
* `ToC` contains a TOC for the page, with links to the item IDs in the main page
* `Body` contains the formatted body of the page
* `Topics` is an ordered list of all of the Topics for the page
* `Keywords` is an ordered list of all of the Keywords for the page
* `Authors` is an ordered list of all of the authors of the page
