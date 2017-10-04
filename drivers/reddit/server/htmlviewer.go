package server

import (
	"database/sql"
	"github.com/coverprice/contentscraper/config"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

// Verify that HtmlViewerRequestHandler implements IRequestHandler interface
var _ IRequestHandler = &HtmlViewerRequestHandler{}

type HtmlViewerRequestHandler struct {
	dbconn *sql.DB
}

func NewHtmlViewerRequestHandler(dbconn *sql.DB) *HtmlViewerRequestHandler {
	return &HtmlViewerRequestHandler{
		dbconn: dbconn,
	}
}

var htmlTemplateStr = `
    <!DOCTYPE html>
    <html>
        <head>
            <meta charset="UTF-8">
            <title>RedditFeed: {{.Name}}</title>
        </head>
        <body>
            <h1>RedditFeed: {{.Name}}</h1>
            <p>
                {{.Description}} - Media: {{.Media}}
            </p>
        </body>
    </html>
`
var htmlTempl = template.Must(template.New("html").Parse(htmlTemplateStr))

func (this *HtmlViewerRequestHandler) HandleFeed(
	feed *config.RedditFeed,
	pagenum int,
	w http.ResponseWriter,
) {
	/*
		data := struct {
			Name string
			Body  string
		}{
			Title: fmt.Sprintf("Feed: %s", feed.Name),
			Body:  "Dummy body",
		}
	*/
	err := htmlTempl.Execute(w, feed)
	if err != nil {
		log.Fatal("Failed to execute html template:", err)
	}
}
