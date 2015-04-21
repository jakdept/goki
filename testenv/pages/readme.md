
This wiki is powered by Javascript, and largely by [MDwiki](http://mdwiki.info) with a [Gitolite](http://gitolite.com/gitolite/index.html) backend. These pages are currently being served to you by [Nginx](http://wiki.nginx.org/Main).

Wiki Structure
--------------

Wiki articles are broken up into folders for categories - for now. Eventually, we will likely try to use keywords,  but for now we'll just be using foldlers and sorting markdown pages into categories.

Reference pages on how this works:

http://dynalon.github.io/mdwiki/#!quickstart.md

https://help.github.com/articles/github-flavored-markdown/

https://github.com/adam-p/markdown-here/wiki/Markdown-Cheatsheet

http://dynalon.github.io/mdwiki/#!gimmicks.md

More functionality will be added as time passes, and reference information will be written.

Currently Planned Categories
--------------

* MySQL
* OpenSSH
* Man Pages
* Exim
* Apache
* Scripts

- - - -

Wiki Syntax
---------

If you read through this wiki, you may notice that it seems spartan. This wiki is currently being written to be very terse. That's intentional (at this point in time).

Wiki Contributions
------------------

If you would like to contribute to this wiki, you may clone it down with:

```
git clone git@colo.perboner.com:lw-mdwiki
```

You may then submit pull requests as desired. Additions will be considered.

Navagation Page
---------------

There is a page at `navagation.md` - this page generates the header for all pages.

- - - - 

Locked Pages
---------------

The following pages are locked in the repository to maintain the wiki's structure. Any commits with unauthorized changes in them will be rejected.

* `index.html`
* `mdwiki.html`
* `index.md`
* `readme.md`
* `navagation.md`

Contributing
------------

If you would like to contribute in the development of this wiki, we welcome you. We are writing pages to try to replace the Liquid Web wiki - as such, we will need a lot of help. However, quality of contributions is vital.

Email Jack Hayhurst if you would like to consider contributing.

Planned Features
----------------

The roadmap for new features includes:

* Code Block Syntax highlighting
 * We need support for more languages
 * Currently, if an invalid language formatting is provided, the page fails to load - there needs to be a fallthrough
* Hidden sections through gimmicks
* Transclusion through iframe gimmick
 * Currently works, but needs work
 * Currently we cannot transclude just one section - filtering needs to be added
 * Relative paths within an included section are translated from the parent page, instead they should be reletive from the included page.
* Category tags through gimmicks
* Automatically generated Category sections
* A Sphinx based search engine
* Include for custom CSS/JS within config.json for static files