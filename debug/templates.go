package debug

import (
	"fmt"
	"net/http"
)

// writeIndex writes the trace index page template to the ResponseWriter.
func writeIndex(w http.ResponseWriter, m Metadata) error {
	return renderIndex(w, m)
}

// writeTrace writes the trace details page template to the ResponseWriter.
func writeTrace(w http.ResponseWriter, traceId string, m Metadata) error {
	be := m.GetBaseEntryByTrace(traceId)
	es := m.GetEntriesByTrace(traceId)
	p := Patch{}
	for _, e := range es {
		for k, v := range buildUpdateSvgActions(e, be) {
			p[k] = v
		}
	}
	err := renderTrace(w, m, traceId, p)
	if err != nil {
		return err
	}
	return nil
}

type Patch map[string]Actions
type Actions map[string]string

// updateSvg processes a metadata entry by updating the appropriate elements of the SVG and
// and returning the current version.
func buildUpdateSvgActions(e Entry, be Entry) Patch {
	// reqColor returns the appropriate color for a request based on the response status code.
	reqColor := func(status int) string {
		if status < 400 || status >= 500 {
			return "green"
		} else {
			return "red"
		}
	}
	// resColor returns the appropriate color for a response based on the response status code.
	resColor := func(status int) string {
		if status < 400 {
			return "green"
		} else {
			return "red"
		}
	}

	newRequestActions := func(statusCode int) Actions {
		return Actions{"color": reqColor(e.Response.StatusCode)}
	}
	newResponseActions := func(statusCode int) Actions {
		return Actions{
			"color": resColor(e.Response.StatusCode),
			"text":  fmt.Sprintf("%d", e.Response.StatusCode),
		}
	}

	var rq, sq string
	if e.ServiceName == be.ServiceName {
		rq = fmt.Sprintf(`-> %s %s %s`, e.ServiceName, e.Request.Method, e.Request.Route)
		sq = fmt.Sprintf(`<-- %s %s %s`, e.ServiceName, e.Request.Method, e.Request.Route)
	} else {
		rq = fmt.Sprintf(`%s %s %s -> %s %s %s`, be.ServiceName, be.Request.Method, be.Request.Route, e.ServiceName, e.Request.Method, e.Request.Route)
		sq = fmt.Sprintf(`%s %s %s <-- %s %s %s`, be.ServiceName, be.Request.Method, be.Request.Route, e.ServiceName, e.Request.Method, e.Request.Route)
	}
	sc := e.Response.StatusCode

	p := Patch{}
	p[rq] = newRequestActions(sc)
	p[sq] = newResponseActions(sc)
	return p
}
