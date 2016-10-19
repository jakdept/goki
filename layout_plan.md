Deployment Overview
===================

Handlers
--------

The basic idea is to have two types of handlers for requests - raw handlers, and markdown handlers.
There are three additional handlers for dealing with different searches.

#### `raw` Handler ####

The following are options for `raw` handlers:

* Allowed extensions/Restricted extensions
* Default request
* Directory to serve from

Raw handlers will retrieve the requested file, and return it.
Files outside of the requested directory should not be served.

Further information.

#### `markdown` Handler ####

The following are options for `markdown` handlers:

* Restricted tags
* Default request
* Directory to serve from
* Template to use

Markdown handlers will retrieve the requested file, translate any markdown in the file into html, parse it into the template, and return the response.
Requests for pages with a restricted tag should not be served - a 403 will be returned.

Further information.

#### `field` Handler ###

The `field` handler lists all articles of a given tag with a given value.
It is typically used to list all articles with a specific topic.

An example configuration might be `domain.com/topic`.
You then configure the tag you want to match - in this case likely `topic`.
Anything left in the URL after that location is matched against that tag.
For instance, `domain.com/topic/ponies` would list all pages where `topic` matches `ponies`.

Further information.

#### `fuzzy` Handler ###

The `fuzzy` handler takes given topics, authors, and a search term.
It then throws all articles matching that pair into the template.

Further information.

#### `query` Handler ####

The `query` handler is a secondary search handler, that allows you to query in a format specific to the indexing engine.

Further information.

Site Repository Layout
----------------------

For my primary location for this server, my content repoository has the following layout:

```
repository
	pages/                # markdown pages, the content of the site
	static/               # various static resources are placed in here to be served as is
	static/css/           # all static css
	static/fonts/         # all static fonts
	static/images/        # all static images
	static/js/            # all static javascript resources
	templates/            # location for storage of all templates
	wiki.index/           # index location, should be ignored and untouched by your source control
	config.json						# config file for server
```

It is recommended that you use something like [harpoon](https://github.com/agrison/harpoon) to then listen for git pushes and deploy.

Feature Requests
----------------

[Use the issue tracker](https://git.liquidweb.com/jhayhurst/goki/issues). We will be using that for both feature requests, and for bugs that pop up.