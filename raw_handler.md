topic:handler
Raw Handler
===========
The Raw handler is the alternate handler used for serving raw files.

Configuration
-------------
An example config section for this handler:

```nohighlight
{
	"ServerType": "raw",
	"Prefix": "/raw/",
	"Path": "/var/www/wiki/",
	"Default": "readme.md",
	"Restricted": [
		"ini",
		"json"
	]
}
```

The elements can appear in any order. The entire configuration is in `json` format.

* `ServerType` is always `raw`
* `Prefix` is the URL path to handle. The most specific Prefix path is used. A trailing `/` will be added automatically.
* `Path` - the phyiscal path to the files on the system.
* `Default` - the default file to serve
* `Restricted` - an array of file extensions that cannot appear

When the request is recieved, it is validated. If it is valid, the Prefix is stripped off, the Path is added to the front, and if needed, the Default and Extension are loaded.

With the above configuration, `http://domain/raw/` would load the page that would also reside at `http://domain/raw/readme.md`.

The raw handler does not use a template. Instead, it detects the `Content-Type` of the file and returns that.