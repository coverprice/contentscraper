package htmlutil

import (
	"os"
	"testing"
)

var testTemplateStr = `
{{define "title"}}Home{{end}}
{{define "content"}}
    <ul>
        <li><a href="/aboutme">About {{.MyName}}</a></li>
    </ul>
{{end}}
`

func TestConfigParsing(t *testing.T) {
	testTemplate := ParseTemplate(testTemplateStr)

	data := struct {
		MyName string
		Breadcrumbs
	}{
		MyName: "James",
		Breadcrumbs: []Breadcrumb{
			NewBreadcrumb("Home", "/"),
			NewBreadcrumb("Reddit", "/reddit"),
			NewBreadcrumb("This Feed", "/reddit/thisfeed"),
		},
	}
	RenderTemplate(os.Stdout, testTemplate, data)
}
