package server

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

var (
	port int
)

func init() {
	flag.IntVar(&port, "port", 8080, "Port to listen on")
}

var indexTemplateStr = `
    <!DOCTYPE html>
    <html>
        <head>
            <meta charset="UTF-8">
            <title>{{.Title}}</title>
            <title>Content Scraper</title>
        </head>
        <body>
            <h1>Hello world!</h1>
            <p>
                Hello world
            </p>
        </body>
    </html>
`

var indexTempl = template.Must(template.New("index").Parse(indexTemplateStr))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Content Scraper",
	}

	err := indexTempl.Execute(w, data)
	if err != nil {
		log.Fatal("Failed to execute index.html template:", err)
	}
}

/*
func showfeed(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()       // parse arguments, you have to call this by yourself
	fmt.Println(r.Form) // print form information in server side
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello astaxie!") // send data to client side
}
*/

// Does not return!
func Launch() {
	http.HandleFunc("/", indexHandler)
	// http.HandleFunc("/redditfeed", index_html) // set router

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("Failed to launch web server: ", err)
	}
}
