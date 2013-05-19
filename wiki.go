// Implementation of http://golang.org/doc/articles/wiki/
package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

type Page struct {
	Title string
	Body  string
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

/*
An http.ResponseWriter that replaces occurrences of [[Title]] with HTML links
to that title.
*/
type LinkedResponseWriter struct {
	http.ResponseWriter
}

var articleLinkRegex = regexp.MustCompile(`\[\[([^.\]/]+)\]\]`)
var fullLinkRegex = regexp.MustCompile(`\[\[(http.+?)\]\]`)
var nakedLinkRegex = regexp.MustCompile(`\[\[(.+?)\]\]`)

func (l *LinkedResponseWriter) Write(p []byte) (int, error) {
	text := string(p)
	text = articleLinkRegex.ReplaceAllString(text, `<a href="/view/$1">$1</a>`)
	text = fullLinkRegex.ReplaceAllString(text, `<a href="$1">$1</a>`)
	text = nakedLinkRegex.ReplaceAllString(text, `<a href="http://$1">$1</a>`)
	return l.ResponseWriter.Write([]byte(text))
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: string(contents)}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(&LinkedResponseWriter{w}, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := &Page{Title: title, Body: r.FormValue("body")}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("templates/edit.html",
	"templates/view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type WikiHandler func(w http.ResponseWriter, r *http.Request, title string)

var titleRegex = regexp.MustCompile(`^[^.\]/]+$`)

func handleWithPrefix(pattern string, handler WikiHandler) {
	validator := func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path
		if titleRegex.MatchString(title) {
			handler(w, r, title)
		} else {
			// "Not Found" here is actually "Forbidden" without having to give
			// explanations. Surprisingly, this is the correct usage per
			// RFC2616.
			http.NotFound(w, r)
		}
	}

	http.Handle(pattern, http.StripPrefix(pattern, http.HandlerFunc(validator)))
}

func main() {
	handleWithPrefix("/view/", viewHandler)
	handleWithPrefix("/edit/", editHandler)
	handleWithPrefix("/save/", saveHandler)
	http.Handle("/", http.RedirectHandler("/view/FrontPage", http.StatusFound))
	http.ListenAndServe("localhost:8080", nil)
}
