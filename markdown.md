topic : touchup

Markdown Formatting
===================

Note: If you are already familiar with Markdown, you might want to look at [the source of this wiki](link to source), which is itself rendered with Gnosis.

Markdown
--------
Gnosis uses Markdown as a markup language.

A simple page could look like:

```
Page Title
==========

SubHeading
----------

  * list item 1
  * list item 2

  This is a hyperlink to [Google](http://google.com).

  Images are like hyperlinks, but with an exclamation mark in front of them:
  ![](http://placekitten.com/g/250/250)

```

Gnosis uses the [GitHub flavored markdown dialect](https://help.github.com/articles/github-flavored-markdown/), so you can i.e. add tables:

    | Tables        | Are           | Cool  |
    | ------------- |:-------------:| -----:|
    | col 3 is      | right-aligned | $1600 |
    | col 2 is      | centered      |   $12 |
    | zebra stripes | are neat      |    $1 |

| Tables        | Are           | Cool  |
| ------------- |:-------------:| -----:|
| col 3 is      | right-aligned | $1600 |
| col 2 is      | centered      |   $12 |
| zebra stripes | are neat      |    $1 |

See the [GitHub Markdown Cheatsheet](https://help.github.com/articles/github-flavored-markdown/) for detailed GFM reference. This is built upon the [original markdwon standard](http://daringfireball.net/projects/markdown/syntax).

Page Title
----------

The top level header tag is reserved for a page title. There should only be one per page, and there needs to be one on every page. This tag needs to appear at the top of the page - below any meta-data and above any other Markdown content.

You may specify this tag with a two line Header tag:

```nohighlight
Page Title
==========
```

Or with a one line tag:

```nohighlight
#Page Title
```

Page Headings
-------------

2nd through 6th level headings may be used on your page to create headings. As a guide, use 2nd level elements to seperate your page into sections, and shy away from using 5th and 6th level headings.

Every heading will show up in the table of contents on the side of the page.

Headings are created with `#` at the start of the line, one `#` for each level.

```nohighlight
##Second Level
### Third Level ###
#### Fourth Level###
```

The closing `#` are optional, but there can be nothing else in the line.

Headings allow inline elements inside of them.

Lists
-----

Lists are created by starting a line with a `*`. You may indent the list by putting a space in front of the `*`.

```nohighlight
* One fish
* Two fish
 * Red fish
 * Blue fish
```

Quotes
------

Quoting text from another source is easy. Simply begin the line with a `>` and the text on that line is quoted.

```nohighlight
> This is a quote.
> It is a really good quote.
>
> - anonymous
```

notice> Added functionality.

As an additional feature, you may prefix your opening quote symbol with a word. The quote box will then be tagged with the word box, prefixed with the word you gave. This is used for Alert boxes.


Links
-----

Hyperlinks are created with standard markdown syntax. Enclose the display text for the link in `[]` and enclose the hyperlink with `()`.

    [Google](http://www.google.com)

Links can also be relative:

    [Go to download](download.md)

Code Blocks
-----------

Preformatted text (like code) can be added and set apart in pages by:

* Putting four spaces at the start of each line.

    Like this

```nohighlight
    Like this
```

Or by using three backticks at the start and the end.

```
Like this
```

    ```
    Like this
    ```

As an additional feature, you may specify a language for syntax highlighting. If you do so, that language will show up in the HTML tag, prefixed with the word `language`, and you can format this with [highlight.js](https://highlightjs.org).

```go
fmt.Println("This is an example\n")
```

    ```go
    fmt.Println("This is an example\n")
    ```

Images
-------

Images are regularly placed as in standard markdown using the `![alt](href "title")` notation. You can group images together by __not__ putting an empty line in between them.

Example:

    ![](http://placekitten.com/g/1200/300 "A kitten")

    ![](http://placekitten.com/g/550/450 "First of two kittens")
    ![](http://placekitten.com/g/550/450 "Second of two kittens")

    ![](http://placekitten.com/g/400/350)
    ![](http://placekitten.com/g/400/350)
    ![](http://placekitten.com/g/400/350)

Will be rendered as:

![](http://placekitten.com/g/1200/300 "A kitten")

![](http://placekitten.com/g/550/450 "First of two kittens")
![](http://placekitten.com/g/550/450 "Second of two kittens")

![](http://placekitten.com/g/400/350)
![](http://placekitten.com/g/400/350)
![](http://placekitten.com/g/400/350)

### Images as Links

To use an image as a link, use the following syntax:

    [![ImageCaption](path/to/image.png)](http://www.linktarget.com)

    Example:
    [![A kitten](http://placekitten.com/g/400/400)](http://www.placekitten.com)

[![A kitten](http://placekitten.com/g/400/400)](http://www.placekitten.com)