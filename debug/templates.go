package debug

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os/exec"
	"time"
)

//var t = template.Must(template.ParseFiles("templates/index.html", "templates/trace.html"))
var t = template.Must(template.Must(template.
	New("indexPage").Parse(index)).
	New("tracePage").Funcs(template.FuncMap{
		"unescape": unescape,
		"json": toJson,
		"toMs": durationToMs,
		"StatusText": http.StatusText,
	}).Parse(trace))

// durationToMs returns a string representation of the duration in ms to one decimal place.
func durationToMs(d time.Duration) string {
	return fmt.Sprintf("%.1f", float64(d.Nanoseconds()) / 1000000.0)
}

func toJson(arg interface{}) string {
	b, _ := json.MarshalIndent(arg, "", "  ")
	return string(b)
}

// writeIndex writes the trace index page template to the ResponseWriter.
func writeIndex(w http.ResponseWriter, m *Metadata) error {
	err := t.ExecuteTemplate(w, "indexPage", m)
	if err != nil {
		return err
	}
	return nil
}

// writeTrace writes the trace details page template to the ResponseWriter.
func writeTrace(w http.ResponseWriter, e *Entry) error {
	dir := "/Users/ladeo/dev/sysl/pkg/arrai"
	textsArg := "{'GET /foobar','GET /todos/{id}','todosResponse','jsonplaceholder.todosResponse'}"
	cmd := exec.Command("arrai", "run", "svg_demo.arrai", textsArg)
	cmd.Dir = dir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, svg)
	}()

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	type trace struct {
		Entry *Entry
		Svg   string
	}
	err = t.ExecuteTemplate(w, "tracePage", trace{e, string(out)})
	if err != nil {
		return err
	}
	return nil
}

// Unescape returns unescaped HTML for use in a template.
func unescape(s string) template.HTML {
	return template.HTML(s)
}

// svg is a hard-coded sequence diagram for the GET /foobar endpoint.
const svg = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" contentScriptType="application/ecmascript" contentStyleType="text/css" height="320px" preserveAspectRatio="none" style="width:486px;height:320px;" version="1.1" viewBox="0 0 486 320" width="486px" zoomAndPan="magnify"><defs><filter height="300%" id="fw4p2vs6dpi2y" width="300%" x="-1" y="-1"><feGaussianBlur result="blurOut" stdDeviation="2.0"/><feColorMatrix in="blurOut" result="blurOut2" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 .4 0"/><feOffset dx="4.0" dy="4.0" in="blurOut2" result="blurOut3"/><feBlend in="SourceGraphic" in2="blurOut3" mode="normal"/></filter></defs><g><text fill="#000000" font-family="sans-serif" font-size="18" lengthAdjust="spacingAndGlyphs" textLength="205" x="136" y="26.708">simple &lt;- GET /foobar</text><rect fill="#FFFFFF" filter="url(#fw4p2vs6dpi2y)" height="87.3984" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="271" y="160.5156"/><rect fill="#FFFFFF" filter="url(#fw4p2vs6dpi2y)" height="29.1328" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="404" y="189.6484"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 5.0,5.0;" x1="276" x2="276" y1="86.25" y2="265.9141"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 5.0,5.0;" x1="409" x2="409" y1="86.25" y2="265.9141"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="48" x="249" y="82.9482">simple</text><ellipse cx="276" cy="53.9531" fill="#FEFECE" filter="url(#fw4p2vs6dpi2y)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="272,41.9531,278,36.9531,276,41.9531,278,46.9531,272,41.9531" style="stroke: #A80036; stroke-width: 1.0;"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="48" x="249" y="277.9092">simple</text><ellipse cx="276" cy="297.2109" fill="#FEFECE" filter="url(#fw4p2vs6dpi2y)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="272,285.2109,278,280.2109,276,285.2109,278,290.2109,272,285.2109" style="stroke: #A80036; stroke-width: 1.0;"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="114" x="349" y="82.9482">jsonplaceholder</text><ellipse cx="409" cy="53.9531" fill="#FEFECE" filter="url(#fw4p2vs6dpi2y)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="405,41.9531,411,36.9531,409,41.9531,411,46.9531,405,41.9531" style="stroke: #A80036; stroke-width: 1.0;"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="114" x="349" y="277.9092">jsonplaceholder</text><ellipse cx="409" cy="297.2109" fill="#FEFECE" filter="url(#fw4p2vs6dpi2y)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="405,285.2109,411,280.2109,409,285.2109,411,290.2109,405,285.2109" style="stroke: #A80036; stroke-width: 1.0;"/><rect fill="#FFFFFF" filter="url(#fw4p2vs6dpi2y)" height="87.3984" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="271" y="160.5156"/><rect fill="#FFFFFF" filter="url(#fw4p2vs6dpi2y)" height="29.1328" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="404" y="189.6484"/><rect fill="#EEEEEE" filter="url(#fw4p2vs6dpi2y)" height="3" style="stroke: #EEEEEE; stroke-width: 1.0;" width="471" x="3" y="116.8164"/><line style="stroke: #000000; stroke-width: 1.0;" x1="3" x2="474" y1="116.8164" y2="116.8164"/><line style="stroke: #000000; stroke-width: 1.0;" x1="3" x2="474" y1="119.8164" y2="119.8164"/><rect fill="#EEEEEE" filter="url(#fw4p2vs6dpi2y)" height="23.1328" style="stroke: #000000; stroke-width: 2.0;" width="186" x="145.5" y="106.25"/><text fill="#000000" font-family="sans-serif" font-size="13" font-weight="bold" lengthAdjust="spacingAndGlyphs" textLength="167" x="151.5" y="122.3169">simple &lt;- GET /foobar</text><polygon fill="#A80036" points="259,156.5156,269,160.5156,259,164.5156,263,160.5156" style="stroke: #A80036; stroke-width: 1.0;"/><line style="stroke: #A80036; stroke-width: 1.0;" x1="3" x2="265" y1="160.5156" y2="160.5156"/><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="76" x="10" y="155.4497">GET /foobar</text><polygon fill="#A80036" points="392,185.6484,402,189.6484,392,193.6484,396,189.6484" style="stroke: #A80036; stroke-width: 1.0;"/><line style="stroke: #A80036; stroke-width: 1.0;" x1="281" x2="398" y1="189.6484" y2="189.6484"/><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="103" x="288" y="184.5825">GET /todos/{id}</text><polygon fill="#A80036" points="292,214.7813,282,218.7813,292,222.7813,288,218.7813" style="stroke: #A80036; stroke-width: 1.0;"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 2.0,2.0;" x1="286" x2="408" y1="218.7813" y2="218.7813"/><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="99" x="298" y="213.7153">todosResponse</text><polygon fill="#A80036" points="14,243.9141,4,247.9141,14,251.9141,10,247.9141" style="stroke: #A80036; stroke-width: 1.0;"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 2.0,2.0;" x1="8" x2="275" y1="247.9141" y2="247.9141"/><text fill="#0000FF" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="203" x="20" y="242.8481">jsonplaceholder.todosResponse</text><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="10" x="227" y="242.8481">&lt;</text><text fill="#008000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="22" x="237" y="242.8481">?, ?</text><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="10" x="259" y="242.8481">&gt;</text></g></svg>`

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
	<style>
		.status {
			color: red;
		}
		.status.OK {
			color: green;
		}
	</style>
</head>
<body>
<p><a href="/-/trace">Back</a></p>

<h1>Trace Details</h1>

<p>Trace ID: {{ .Entry.TraceId }}</p>
<p>
	Status:
	<span class="status {{ StatusText .Entry.Status }}">
		{{ .Entry.Status }} {{ StatusText .Entry.Status }}
	</span>
	({{ toMs .Entry.Latency }}ms)
</p>

{{ unescape .Svg }}
<h2>Request</h2>
<pre>{{ json .Entry.Request }}</pre>

<h2>Response</h2>
<pre>{{ .Entry.Response }}</pre>

</body>
</html>`
