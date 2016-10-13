topic: index
topic: config
Configuration
=============

The configuration file for this server is [JSON based](http://json.org/).
The major item organization is:

```go
type Config struct {
	Redirects []RedirectSection
	Indexes   []IndexSection{
    Server    []ServerSection
  }
}
```

Of note, you have to include `{}` curley braces around each of the server configurations.

Each section within the main section will have it's own array.
Further, the items with a `[]` in front of them can have multiple sections.

GlobalSection
-------------

The GlobalSection is the outermost element, and has the following structure:

```go
type GlobalSection struct {
	Address     string
	Port        string
	Hostname    string
	TemplateDir string
	CertFile    string
	KeyFile     string
	Indexes     []IndexSection
	Redirects   []RedirectSection
}
```

An example `json` configuration:

```json
"Global": {
  "Address":"*",
  "Port": "8080",
  "Hostname": "localhost",
  "TemplateDir": "templates/"
},
```

* `Address` is the address to listen on
* `Port` is the port that you want the server to listen on
* `Hostname` is the hostname the server should respond with
* `TemplateDir` is a directory containing all of the templates, and nothing else
* `CertFile` is the file containing the certificate
* `KeyFile` is the file containing the SSL keyfile

RedirectSection
---------------

notice> A `RedirectSection` may not redirect the root of a handler.

Each `RedirectSection` represents a unique redirect to run on the server.

Redirects are loaded into a `go` structure of:

```go
type RedirectSection struct {
	Requested string
	Target    string
	Code      int
}
```

An example redirect:

```json
"Redirects": [
	{
	  "Requested": "/favicon.ico",
	  "Target": "/site/favicon.ico",
	  "Code": 301
	}
]
```

* `Requested` is the path to watch for - when a request for that path is recieved, action is taken.
* `Target` is the location to point the redirect at.
* `Code` is the HTTP response code to use.

You may have multiple redirects - an example of this would look like:

```json
"Redirects": [
  {
    "Requested": "/favicon.ico",
    "Target": "/site/favicon.ico",
    "Code": 301
  },
  {
    "Requested": "/rawindex",
    "Target": "/raw/index.md",
    "Code": 307
  }
]
```

IndexSection
------------

Each `IndexSection` represents a unique index to create.

Each Index Section is loaded into a `go` structure of the type:

```go
type IndexSection struct {
	WatchDirs      map[string]string
	WatchExtension string
	IndexPath      string
	IndexType      string
	IndexName      string
	Restricted     []string
	Handlers       []ServerSection
}
```

A example `json` index section:

```json
"Indexes": [
  {
    "WatchDirs": {
      "./wiki/": "/"
    },
    "WatchExtension": ".md",
    "IndexType": "en",
    "IndexPath": "./junkindex/wiki.index",
    "IndexName": "wiki",
    "Restricted": [
      "internal",
      "handbookt"
    ],
    "Handlers":[]
  }
] 
```

* `WatchDirs` is a list of directories to index, mapped to the URI to write them with
* `WatchExtension` is the file extension to add - only files with that extension will be added
* `IndexType` specifies the language filter to use when indexing
* `IndexPath` is the location to put the index on the disk
* `IndexName` specifies the name to give to the index
* `Restricted` is a list of topics within pages to not index
* `Handlers` contains the handlers that run under that index for the server

Each file found within a `WatchDir` will be stored with the `URIPath` set as the path to that file, minus the first part of the `WatchDir`, prepended with the second.

When indexing, if a page contains a `topic` that is in the `Restricted` list, that page will not be indexed.

You may create multiple distinct indexes:

```json
"Indexes": [
  {
    "WatchDirs": {
      "./wiki/": "/"
    },
    "WatchExtension": ".md",
    "IndexPath": "./junkindex/wiki.index",
    "IndexType": "en",
    "IndexName": "wiki",
    "Restricted": [
      "internal",
      "handbookt"
    ],
    "Handlers":[]
  },
  {
    "WatchDirs": {
      "./otherwiki/": "/"
    },
    "WatchExtension": ".md",
    "IndexPath": "./someother/wiki.index",
    "IndexType": "en",
    "IndexName": "wiki",
    "Restricted": [
      "internal",
      "handbookt"
    ],
    "Handlers":[]
  }
] 
```

You may also map multiple directories in one index:


```json
"Indexes": [
  {
    "WatchDirs": {
      "./wiki/": "/",
      "./otherwiki/": "/other/"
    },
    "WatchExtension": ".md",
    "IndexPath": "./junkindex/wiki.index",
    "IndexType": "en",
    "IndexName": "wiki",
    "Restricted": [
      "internal",
      "handbookt"
    ],
    "Handlers":[]
  }
] 
```

ServerSection
-------------

`ServerSection` is used to specify the different `routes`, `handlers`, or `servers` to use to process different requests. You may have as many as you need.

```go
type ServerSection struct {
	Path               string
	Prefix             string
	Default            string
	Template           string
	FallbackTemplate   string
	ServerType         string
	TopicURL           string
	Restricted         []string
}
```

The `ServerType` value specifies which server type to use. Each different `ServerType` has it's own page of documentation:

* `raw` is [documented in raw_handler.md](raw_handler.md)
* `markdown` is [documented in markdown_handler.md](markdown_handler.md)
* `fieldList` is [documented in fieldlist_handler.md](fieldlist_handler.md)