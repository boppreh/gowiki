// Implementation of http://golang.org/doc/articles/wiki/
package main

import (
    "fmt"
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

var titleRegex = "[a-zA-Z0-9 ]+"
var linkRegex = regexp.MustCompile(`\[\[` + titleRegex + `\]\]`)

func (l *LinkedResponseWriter) Write(p []byte) (int, error) {
    linkReplacer := func(title []byte) []byte {
        // Remove [[ and ]]
        strTitle := string(title)[2:len(title)-2]
        return []byte(fmt.Sprintf(`<a href="/view/%v">%v</a>`,
                                  strTitle, strTitle))
    }
    return l.ResponseWriter.Write(linkRegex.ReplaceAllFunc(p, linkReplacer))
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
	body := r.FormValue("body")
	p := &Page{Title: title, Body: body}
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

var titleValidator = regexp.MustCompile("^" + titleRegex + "$")

type WikiHandler func (w http.ResponseWriter, r *http.Request, title string)

func handleWithPrefix(pattern string, handler WikiHandler) {
	validator := func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path
		if titleValidator.MatchString(title) {
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
