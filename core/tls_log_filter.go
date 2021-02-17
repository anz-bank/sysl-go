package core

import (
	"regexp"

	"github.com/anz-bank/sysl-go/log"
)

type TLSLogFilter struct {
	logger log.Logger
	re     *regexp.Regexp
}

func (t *TLSLogFilter) Write(p []byte) (n int, err error) {
	if t.re.Match(p) {
		t.logger.Debug(string(p))
	} else {
		t.logger.Info(string(p))
	}

	return len(p), nil
}
