# Web render

This module implements a static renderer for the HTML, CSS and SVG formats.

It consists for the main part of a Golang port of the awesome [Weasyprint](https://github.com/Kozea/WeasyPrint) python Html to Pdf library.

This is an **ongoing work**, not production ready just yet.

## Scope

The main goal of this module is to process HTML or SVG inputs into laid out documents, ready to be paint, and to be compatible with various output formats (like raster images or PDF files).
To do so, this module uses an abstraction of the output, whose implementation must be provided by an higher level package.

## Outline of the module

From the lower level to the higher level, this module has the following structure :

- the `css` package provides a CSS parser, with property validation and a CSS selector engine (`css/selector`).

- the `svg` package implements a SVG parser and renderer, supporting CSS styling.

- the `html` package implements an HTML renderer

- the `backend` package defines the interfaces which must be implemented by output targets.

The main entry points are the `html/document` package for HTML rendering and the `svg` package if you only need SVG support.

### HTML to PDF: an overview

The `html` package implements a static HTML renderer, which works by :

- parsing the HTML input and fetching CSS files, and cascading the styles. This is implemented in the `html/tree` package

- building a tree of boxes from the HTML structure (package `html/boxes`)

- laying out this tree, that is attributing position and dimensions to the boxes, and performing line, paragraph and page breaks (package `html/layout`)

- drawing the laid out tree to an output. Contrary to the Python library, this step is here performed on an abstract output, which must implement the `backend.Document` interface. This means than the core layout logic could easily be reused for other purposes, such as visualizing html document on a GUI application, or targetting other output file formats.
