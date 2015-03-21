Deployment Overview
===================

Handlers
--------

The initial plan is to have two types of handlers for requests - raw handlers, and markdown handlers.

#### Raw Handler ####

The following options will be available for raw handlers:

* Allowed extensions/Restricted extensions
* Default request
* Directory to serve from
* Restricted folders

Raw handlers will retrieve the requested file, and return it. Files outside of the requested directory should not be served.

#### Markdown Handler ####

The following options will be available for raw handlers:

* Restricted tags
* Default request
* Directory to serve from
* Restricted folders
* Template to use

Raw handlers will retrieve the requested file, translate any markdown in the file into html, parse it into the template, and return the response. Requests for pages with a restricted tag should not be served - an alternate 403 page will be served.

Content Repositories
--------------------

To facilitate this project, we will use multiple Git repositories. At this point in time, the plan is just to install gitolite on the servers, and use stuff within them for automatic pushing.

* One repository will contain the actual content of the wiki - markdown pages, required images, etc.
* One repository will contain the site files - templates, css, javascript, config for the server
 * There will be one version of this respository for the internal servers - allows downloading of `.md` files and no restricted tags
 * There will be a second version of this respository for the external servers - do not allow downloading of `.md` files, do not allow viewing of pages with restricted tags

Site Repository Layout
----------------------

For the site repositories, the plan is to have the following layout:

```
repository
	public/								# this directory will be served through a raw handler
	public/css/						# any site css
	public/javascript/		# any site javascript
	public/static/				# any site static files (403 pages, 404 pages, etc)
	template/							# any site templates
	githooks/							# git hook scripts for deployment
	system/								# contains firewall configuration, systemd config
	system/keys/					# public SSH keys for push access to git repositories/servers
	system/bin/						# contains sysv init script, program binaries, install script, update script
	config.json						# config file for server
```

For the content repository, it doesn't really matter what the layout is.

We could create folders, and then lock down edits to files within those folders to just certain departments - for instance, create a secteam+esc folder and only give secteam and escalations permission to edit it, or create just a secteam folder and put the known RBL issues in there?

Feature Requests
----------------

[Use the issue tracker](https://git.liquidweb.com/jhayhurst/gnosis/issues). We will be using that for both feature requests, and for bugs that pop up.