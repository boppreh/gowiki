gowiki
======

`gowiki` is a very simple Wiki server written in Go, with features to
view, edit and link pages. All pages are stored in disk (`data/`) and
displayed with templates (`templates/`). The core code is from the great
tutorial at http://golang.org/doc/articles/wiki/ .

Features
--------

- Article titles can contain spaces and unicode characters
- Links automatically added to mentions in format `[[title]]`
- Format `[[text]]` makes an absolute link if it's not a valid title
- Links to inexistent articles go straight to the edit page
- Edit and view pages are rendered from HTML templates (dir `templates/`)
- Templates and data have their own separate folders
- Root resource redirects to article `FrontPage`
- Page loads and saves are all done directly to the disk
