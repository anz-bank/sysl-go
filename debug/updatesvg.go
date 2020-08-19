package debug

import (
	"encoding/json"
	"fmt"
	"github.com/arr-ai/arrai/syntax"
	"github.com/sirupsen/logrus"
	"strings"
)

// updateSvg returns a copy of SVG with color applies to text and its associated elements.
func updateSvg(svg string, p Patch) (string, error) {
	b := strings.Builder{}
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	err := e.Encode(p)
	if err != nil {
		return "", err
	}

	call := fmt.Sprintf("apply(`%s`, %s)", svg, b.String())
	logrus.Info(call)
	value, err := syntax.EvaluateExpr("", fmt.Sprintf("%s.%s", updateScript, call))
	if err != nil {
		return "", err
	}

	return value.String(), nil
}
