package debug

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"html/template"
	"io"
	"net/http"
	"os/exec"
	"time"
)

var funcs = template.FuncMap{
	"unescape":   unescape,
	"json":       toJson,
	"toMs":       durationToMs,
	"StatusText": http.StatusText,
}

//var t = template.Must(template.ParseFiles("templates/index.html", "templates/trace.html"))
var t = template.Must(template.Must(template.Must(template.
	New("indexPage").Funcs(funcs).Parse(index)).
	New("tracePage").Funcs(funcs).Parse(trace)).
	New("subtrace").Funcs(funcs).Parse(subtrace))

// durationToMs returns a string representation of the duration in ms to one decimal place.
func durationToMs(d time.Duration) string {
	return fmt.Sprintf("%.1f", float64(d.Nanoseconds())/1000000.0)
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

func apply(svg string, e Entry, be Entry) (string, string, error) {
	currentSvg := svg
	var textsArg string
	var err error

	if e.ServiceName == be.ServiceName {
		textsArg = fmt.Sprintf(`{'-> %s %s %s'}`, e.ServiceName, e.Request.Method, e.Request.Route)
		var colorArg string
		if e.Response.Status < 400 || e.Response.Status >= 500 {
			colorArg = "green"
		} else {
			colorArg = "red"
		}
		currentSvg, err = updateSvg(currentSvg, textsArg, colorArg)
		if err != nil {
			logrus.WithError(err).Error("upstream inbound svg update failed")
		}

		textsArg = fmt.Sprintf(`{'<-- %s %s %s'}`, e.ServiceName, e.Request.Method, e.Request.Route)
		if e.Response.Status < 400 {
			colorArg = "green"
		} else {
			colorArg = "red"
		}
		currentSvg, err = updateSvg(currentSvg, textsArg, colorArg)
		if err != nil {
			logrus.WithError(err).Error("upstream outbound svg update failed")
		}
	} else {
		textsArg = fmt.Sprintf(`{'%s %s %s -> %s %s %s'}`, be.ServiceName, be.Request.Method, be.Request.Route, e.ServiceName, e.Request.Method, e.Request.Route)
		var colorArg string
		if e.Response.Status < 400 || e.Response.Status >= 500 {
			colorArg = "green"
		} else {
			colorArg = "red"
		}
		currentSvg, err = updateSvg(currentSvg, textsArg, colorArg)
		if err != nil {
			logrus.WithError(err).Error("downstream outbound svg update failed")
		}

		textsArg = fmt.Sprintf(`{'%s %s %s <-- %s %s %s'}`, be.ServiceName, be.Request.Method, be.Request.Route, e.ServiceName, e.Request.Method, e.Request.Route)
		if e.Response.Status < 400 {
			colorArg = "green"
		} else {
			colorArg = "red"
		}
		currentSvg, err = updateSvg(currentSvg, textsArg, colorArg)
		if err != nil {
			logrus.WithError(err).Error("downstream inbound svg update failed")
		}
	}

	// Check entry to determine success/failure
	// Update svg based on response

	// Render subtemplate with entry data
	var sub bytes.Buffer
	err = t.ExecuteTemplate(&sub, "subtrace", e)
	if err != nil {
		logrus.WithError(err).Error("subtemplate render failed")
	}
	result := sub.String()

	// Return updated svg and subtemplate
	return currentSvg, result, err
}

// updateSvg invokes the arr.ai script to update svg with color.
func updateSvg(svg, text, color string) (string, error) {
	cmd := exec.Command("arrai", "run", "svg_demo.arrai", text, color)
	cmd.Dir = "/Users/ladeo/dev/sysl/pkg/arrai"

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return svg, err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, svg)
	}()

	var errb bytes.Buffer
	cmd.Stderr = &errb

	out, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).Error(errb.String())
		return svg, err
	}
	return string(out), nil
}

// writeTrace writes the trace details page template to the ResponseWriter.
func writeTrace(w http.ResponseWriter, traceId string, metadata *Metadata) error {
	be := metadata.GetBaseEntryByTrace(traceId)
	es := metadata.GetEntriesByTrace(traceId)

	subs := []string{}
	currentSvg := svg
	for _, e := range es {
		newSvg, sub, err := apply(currentSvg, e, be)
		currentSvg = newSvg
		subs = append(subs, sub)
		if err != nil {
			logrus.Error(err)
		}
	}

	type trace struct {
		//Entry *Entry
		Subs    []string
		Svg     string
		TraceId string
	}
	err := t.ExecuteTemplate(w, "tracePage", trace{subs, currentSvg, traceId})
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
const svg = `<?xml version="1.0" encoding="UTF-8" standalone="no"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" contentScriptType="application/ecmascript" contentStyleType="text/css" height="333px" preserveAspectRatio="none" style="width:263px;height:333px;" version="1.1" viewBox="0 0 263 333" width="263px" zoomAndPan="magnify"><defs><filter height="300%" id="fciq4dyk34bhh" width="300%" x="-1" y="-1"><feGaussianBlur result="blurOut" stdDeviation="2.0"/><feColorMatrix in="blurOut" result="blurOut2" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 .4 0"/><feOffset dx="4.0" dy="4.0" in="blurOut2" result="blurOut3"/><feBlend in="SourceGraphic" in2="blurOut3" mode="normal"/></filter></defs><g><rect fill="#FFFFFF" filter="url(#fciq4dyk34bhh)" height="29.1328" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="20" y="125.5625"/><rect fill="#FFFFFF" filter="url(#fciq4dyk34bhh)" height="29.1328" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="89.5" y="183.8281"/><rect fill="#FFFFFF" filter="url(#fciq4dyk34bhh)" height="145.6641" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="197" y="96.4297"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 5.0,5.0;" x1="25" x2="25" y1="65.2969" y2="260.0938"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 5.0,5.0;" x1="94" x2="94" y1="65.2969" y2="260.0938"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 5.0,5.0;" x1="202" x2="202" y1="65.2969" y2="260.0938"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="28" x="8" y="61.9951">Visa</text><ellipse cx="25" cy="33" fill="#FEFECE" filter="url(#fciq4dyk34bhh)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="21,21,27,16,25,21,27,26,21,21" style="stroke: #A80036; stroke-width: 1.0;"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="28" x="8" y="272.0889">Visa</text><ellipse cx="25" cy="291.3906" fill="#FEFECE" filter="url(#fciq4dyk34bhh)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="21,279.3906,27,274.3906,25,279.3906,27,284.3906,21,279.3906" style="stroke: #A80036; stroke-width: 1.0;"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="79" x="52" y="61.9951">MasterCard</text><ellipse cx="94.5" cy="33" fill="#FEFECE" filter="url(#fciq4dyk34bhh)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="90.5,21,96.5,16,94.5,21,96.5,26,90.5,21" style="stroke: #A80036; stroke-width: 1.0;"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="79" x="52" y="272.0889">MasterCard</text><ellipse cx="94.5" cy="291.3906" fill="#FEFECE" filter="url(#fciq4dyk34bhh)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="90.5,279.3906,96.5,274.3906,94.5,279.3906,96.5,284.3906,90.5,279.3906" style="stroke: #A80036; stroke-width: 1.0;"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="104" x="147" y="61.9951">PaymentServer</text><path d="M184,13 C184,3 202,3 202,3 C202,3 220,3 220,13 L220,39 C220,49 202,49 202,49 C202,49 184,49 184,39 L184,13 " fill="#FEFECE" filter="url(#fciq4dyk34bhh)" style="stroke: #000000; stroke-width: 1.5;"/><path d="M184,13 C184,23 202,23 202,23 C202,23 220,23 220,13 " fill="none" style="stroke: #000000; stroke-width: 1.5;"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="104" x="147" y="272.0889">PaymentServer</text><path d="M184,285.3906 C184,275.3906 202,275.3906 202,275.3906 C202,275.3906 220,275.3906 220,285.3906 L220,311.3906 C220,321.3906 202,321.3906 202,321.3906 C202,321.3906 184,321.3906 184,311.3906 L184,285.3906 " fill="#FEFECE" filter="url(#fciq4dyk34bhh)" style="stroke: #000000; stroke-width: 1.5;"/><path d="M184,285.3906 C184,295.3906 202,295.3906 202,295.3906 C202,295.3906 220,295.3906 220,285.3906 " fill="none" style="stroke: #000000; stroke-width: 1.5;"/><rect fill="#FFFFFF" filter="url(#fciq4dyk34bhh)" height="29.1328" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="20" y="125.5625"/><rect fill="#FFFFFF" filter="url(#fciq4dyk34bhh)" height="29.1328" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="89.5" y="183.8281"/><rect fill="#FFFFFF" filter="url(#fciq4dyk34bhh)" height="145.6641" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="197" y="96.4297"/><polygon fill="#808080" points="185,92.4297,195,96.4297,185,100.4297,189,96.4297" style="stroke: #808080; stroke-width: 1.0;"/><line style="stroke: #808080; stroke-width: 1.0;" x1="3" x2="191" y1="96.4297" y2="96.4297"/><a href="#%20%20-&gt;%20PaymentServer%20POST%20/pay" target="_top" title="#%20%20-&gt;%20PaymentServer%20POST%20/pay" xlink:actuate="onRequest" xlink:href="#%20%20-&gt;%20PaymentServer%20POST%20/pay" xlink:show="new" xlink:title="#%20%20-&gt;%20PaymentServer%20POST%20/pay" xlink:type="simple"><text fill="#0000FF" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" text-decoration="underline" textLength="65" x="10" y="91.3638">POST /pay</text></a><polygon fill="#808080" points="41,121.5625,31,125.5625,41,129.5625,37,125.5625" style="stroke: #808080; stroke-width: 1.0;"/><line style="stroke: #808080; stroke-width: 1.0;" x1="35" x2="196" y1="125.5625" y2="125.5625"/><a href="#PaymentServer%20POST%20/pay%20-&gt;%20Visa%20POST%20/pay" target="_top" title="#PaymentServer%20POST%20/pay%20-&gt;%20Visa%20POST%20/pay" xlink:actuate="onRequest" xlink:href="#PaymentServer%20POST%20/pay%20-&gt;%20Visa%20POST%20/pay" xlink:show="new" xlink:title="#PaymentServer%20POST%20/pay%20-&gt;%20Visa%20POST%20/pay" xlink:type="simple"><text fill="#0000FF" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" text-decoration="underline" textLength="65" x="47" y="120.4966">POST /pay</text></a><polygon fill="#808080" points="185,150.6953,195,154.6953,185,158.6953,189,154.6953" style="stroke: #808080; stroke-width: 1.0;"/><line style="stroke: #808080; stroke-width: 1.0; stroke-dasharray: 2.0,2.0;" x1="25" x2="191" y1="154.6953" y2="154.6953"/><a href="#PaymentServer%20POST%20/pay%20&lt;--%20Visa%20POST%20/pay" target="_top" title="#PaymentServer%20POST%20/pay%20&lt;--%20Visa%20POST%20/pay" xlink:actuate="onRequest" xlink:href="#PaymentServer%20POST%20/pay%20&lt;--%20Visa%20POST%20/pay" xlink:show="new" xlink:title="#PaymentServer%20POST%20/pay%20&lt;--%20Visa%20POST%20/pay" xlink:type="simple"><text fill="#0000FF" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" text-decoration="underline" textLength="15" x="32" y="149.6294">ok</text></a><polygon fill="#808080" points="110.5,179.8281,100.5,183.8281,110.5,187.8281,106.5,183.8281" style="stroke: #808080; stroke-width: 1.0;"/><line style="stroke: #808080; stroke-width: 1.0;" x1="104.5" x2="196" y1="183.8281" y2="183.8281"/><a href="#PaymentServer%20POST%20/pay%20-&gt;%20MasterCard%20POST%20/pay" target="_top" title="#PaymentServer%20POST%20/pay%20-&gt;%20MasterCard%20POST%20/pay" xlink:actuate="onRequest" xlink:href="#PaymentServer%20POST%20/pay%20-&gt;%20MasterCard%20POST%20/pay" xlink:show="new" xlink:title="#PaymentServer%20POST%20/pay%20-&gt;%20MasterCard%20POST%20/pay" xlink:type="simple"><text fill="#0000FF" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" text-decoration="underline" textLength="65" x="116.5" y="178.7622">POST /pay</text></a><polygon fill="#808080" points="185,208.9609,195,212.9609,185,216.9609,189,212.9609" style="stroke: #808080; stroke-width: 1.0;"/><line style="stroke: #808080; stroke-width: 1.0; stroke-dasharray: 2.0,2.0;" x1="94.5" x2="191" y1="212.9609" y2="212.9609"/><a href="#PaymentServer%20POST%20/pay%20&lt;--%20MasterCard%20POST%20/pay" target="_top" title="#PaymentServer%20POST%20/pay%20&lt;--%20MasterCard%20POST%20/pay" xlink:actuate="onRequest" xlink:href="#PaymentServer%20POST%20/pay%20&lt;--%20MasterCard%20POST%20/pay" xlink:show="new" xlink:title="#PaymentServer%20POST%20/pay%20&lt;--%20MasterCard%20POST%20/pay" xlink:type="simple"><text fill="#0000FF" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" text-decoration="underline" textLength="15" x="101.5" y="207.895">ok</text></a><polygon fill="#808080" points="14,238.0938,4,242.0938,14,246.0938,10,242.0938" style="stroke: #808080; stroke-width: 1.0;"/><line style="stroke: #808080; stroke-width: 1.0; stroke-dasharray: 2.0,2.0;" x1="8" x2="201" y1="242.0938" y2="242.0938"/><a href="#%20%20&lt;--%20PaymentServer%20POST%20/pay" target="_top" title="#%20%20&lt;--%20PaymentServer%20POST%20/pay" xlink:actuate="onRequest" xlink:href="#%20%20&lt;--%20PaymentServer%20POST%20/pay" xlink:show="new" xlink:title="#%20%20&lt;--%20PaymentServer%20POST%20/pay" xlink:type="simple"><text fill="#0000FF" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" text-decoration="underline" textLength="15" x="20" y="237.0278">ok</text></a></g></svg>`

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
{{range .TraceIds}}
<li><a href="/-/trace/{{.}}">{{.}}</a></li>
{{end}}
</ul>
</body>
</html>`

const subtrace = `
<div class="subtrace {{ .Id }}" style="display: none">
<p>{{ .ServiceName }}: {{ .Request.Method }} {{ .Request.Route }} ({{ .Request.URL }})</p>

<h2>Request</h2>
<h3>Headers</h3>
<pre>{{ json .Request.Headers }}</pre>

<h3>Body</h3>
<pre>{{ .Request.Body }}</pre>

<h2>Response</h2>
<p>
	Status:
	<span class="status {{ StatusText .Response.Status }}">
		{{ .Response.Status }} {{ StatusText .Response.Status }}
	</span>
	({{ toMs .Response.Latency }}ms)
</p>

<h3>Headers</h3>
<pre>{{ json .Response.Headers }}</pre>

<h3>Body</h3>
<pre>
{{if .Response.Body}}
{{ .Response.Body }}
{{else}}
(none)
{{end}}
</pre>
</div>
`

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

<p>Trace ID: {{ .TraceId }}</p>

{{ unescape .Svg }}

{{ range .Subs }}
{{ unescape . }}
{{end}}

<script>
function display(href) {
	let [_, lhs, arrow, rhs] = href.replace(/%20/g, ' ').match(/#([^-<>]+)([-<>]+)(.+)/);
	lhs = lhs && lhs.trim();
	rhs = rhs && rhs.trim();
	arrow = arrow && arrow.trim();
	console.log(lhs, arrow, rhs);

	const tc = rhs.replace(/[ /]/g, '_').toLowerCase();
	console.log(tc);
	document.querySelectorAll('div.subtrace').forEach(d => d.setAttribute('style', 'display:none'));
	document.querySelectorAll('div.subtrace.'+tc).forEach(d => d.setAttribute('style', 'display:block'));
}

document.querySelectorAll('a').forEach(
	a => a.addEventListener('mouseover',
		e => {
			display(e.currentTarget.getAttribute('href'))
			document.querySelectorAll('a').forEach(a => a.removeAttribute('style'))
			e.currentTarget.setAttribute('style', 'font-weight: bold');
		}
	)
)
</script>
</body>
</html>`
