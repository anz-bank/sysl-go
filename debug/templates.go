package debug

import (
	"html/template"
	"net/http"
)

//var t = template.Must(template.ParseFiles("templates/index.html", "templates/trace.html"))
var t = template.Must(template.Must(template.
	New("indexPage").Parse(index)).
	New("tracePage").Parse(trace))

func writeIndex(w http.ResponseWriter, m *Metadata) error {
	err := t.ExecuteTemplate(w, "indexPage", m)
	if err != nil {
		return err
	}
	return nil
}

func writeTrace(w http.ResponseWriter, e *Entry) error {
	err := t.ExecuteTemplate(w, "tracePage", e)
	if err != nil {
		return err
	}
	return nil
}

const index = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Debug Traces</title>
</head>
<body>
<h1>Debug Traces</h1>

<ul>
{{range .Entries}}
<li><a href="/-/trace/{{.TraceId}}">{{.TraceId}}</a></li>
{{end}}
</ul>
</body>
</html>`

const trace = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Trace Details</title>
</head>
<body>
<p><a href="/-/trace">Back</a></p>

<h1>Trace Details</h1>

<p>Trace ID: {{.TraceId}}</p>
</body>
</html>`
