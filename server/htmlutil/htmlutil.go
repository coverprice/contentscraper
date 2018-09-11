package htmlutil

import (
	log "github.com/sirupsen/logrus"
	"html/template"
	"io"
	"reflect"
	"strings"
)

type Breadcrumbs []Breadcrumb

type Breadcrumb struct {
	Title string
	Url   string
}

func NewBreadcrumb(title, url string) Breadcrumb {
	return Breadcrumb{
		Title: title,
		Url:   url,
	}
}

var baseTemplateStr = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <link rel="shortcut icon" type="image/png" href="/static/icons8-peach-48.png" />
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-beta/css/bootstrap.min.css" integrity="sha384-/Y6pD6FV/Vv2HJnA6t+vslU6fwYXjCFtcEpHbNJ0lyAFsXTsjBbfaDjzALeQsN6M" crossorigin="anonymous">
    {{block "style" .}} {{end}}
    <title>{{block "title" .}}[Default title]{{end}}</title>
  </head>
  <body>
    {{block "breadcrumb" .}}
        {{if hasField . "Breadcrumbs"}}
            <ol class="breadcrumb">
            {{range .Breadcrumbs}}
                <li class="breadcrumb-item"><a href="{{.Url}}#">{{.Title}}</a></li>
            {{end}}
            </ol>
        {{end}}
    {{end}}

    {{block "content" .}}{{end}}

    <!-- Optional JavaScript -->
    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
    <script src="https://code.jquery.com/jquery-3.2.1.slim.min.js" integrity="sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.11.0/umd/popper.min.js" integrity="sha384-b/U6ypiBEHpOf/4+1nzFpr53nxSS+GLCkfwBdFNTxtclqqenISfwAzpKaMNFNmj4" crossorigin="anonymous"></script>
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-beta/js/bootstrap.min.js" integrity="sha384-h0AbiXch4ZDo7tp9hKZ4TsHbi047NrKGLO3SEJAg45jXxnGIfYzk4Si90RDIqNm1" crossorigin="anonymous"></script>
    {{block "js" .}}{{end}}
  </body>
</html>
`

// Given an arbitrary value, determines if it's a Struct (or a pointer to a Struct), and returns
// whether that Struct has a field with the given name.
func hasField(v interface{}, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	return rv.FieldByName(name).IsValid()
}

// Returns whether the given string ends with the given suffix
func hasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Allows the above functions to be called from within the template.
var (
	baseTemplate = template.Must(
		template.New("base").Funcs(
			template.FuncMap{
				"hasField":  hasField,
				"hasSuffix": hasSuffix,
			},
		).Parse(baseTemplateStr),
	)
)

func ParseTemplate(templateStr string) *template.Template {
	t, err := template.Must(baseTemplate.Clone()).Parse(templateStr)
	if err != nil {
		log.Fatal("Could not parse template", err)
	}
	return t
}

func RenderTemplate(wr io.Writer, t *template.Template, data interface{}) {
	if err := t.Execute(wr, data); err != nil {
		log.Fatal("Could not render template", err)
	}
}
