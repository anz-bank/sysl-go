package debug

import (
	"encoding/json"
	"fmt"
	"github.com/arr-ai/arrai/syntax"
	"github.com/sirupsen/logrus"
	"io"
)

//go:generate ./generate.render.arrai.go.sh

// renderIndex returns the rendered trace index page.
func renderIndex(w io.Writer, m Metadata) error {
	j, err := json.Marshal(m)
	if err != nil {
		return err
	}

	call := fmt.Sprintf("index(`%s`)", j)
	logrus.Infof(call)
	value, err := syntax.EvaluateExpr(
		"", fmt.Sprintf("%s.%s", renderScript, call))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(value.String()))
	if err != nil {
		return err
	}
	return nil
}

// renderTrace returns the rendered trace details page.
func renderTrace(w io.Writer, m Metadata, traceId, svg string) error {
	j, err := json.Marshal(m)
	if err != nil {
		return err
	}

	call := fmt.Sprintf("trace(`%s`, `%s`, `%s`)", j, traceId, svg)
	logrus.Infof(call)
	value, err := syntax.EvaluateExpr("", fmt.Sprintf("%s.%s", renderScript, call))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(value.String()))
	if err != nil {
		return err
	}
	return nil
}
