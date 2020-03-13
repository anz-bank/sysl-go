package core

import (
	"regexp"

	"github.com/sirupsen/logrus"
)

type TLSLogFilter struct {
	logger *logrus.Logger
	re     *regexp.Regexp
}

func (t *TLSLogFilter) Write(p []byte) (n int, err error) {
	level := logrus.WarnLevel
	if t.re.Match(p) {
		level = logrus.DebugLevel
	}
	t.logger.Log(level, string(p))

	return len(p), nil
}
