package logging

import (
	"fmt"
	"io"
	"strings"

	"github.com/anz-bank/sysl-go-comms/config"
	"github.com/anz-bank/sysl-go-comms/splunk"
	"github.com/sirupsen/logrus"
)

func LogFormatter(format string) (logrus.Formatter, error) {
	lf := strings.ToLower(format)
	if lf == "text" {
		return &logrus.TextFormatter{DisableColors: true}, nil
	}
	if lf == "color" {
		return &logrus.TextFormatter{DisableColors: false}, nil
	}
	if lf == "json" {
		return &logrus.JSONFormatter{}, nil
	}
	return nil, fmt.Errorf("unknown log format '%s' [color, text, json]", format)
}

func Logger(w io.Writer, cfg *config.LogConfig) (*logrus.Logger, error) {
	logFormatter, err := LogFormatter(cfg.Format)
	if err != nil {
		return nil, err
	}
	logger := &logrus.Logger{
		Out:          w,
		Formatter:    logFormatter,
		Hooks:        make(logrus.LevelHooks),
		Level:        cfg.Level,
		ReportCaller: cfg.ReportCaller,
	}

	if cfg.Splunk != nil {
		httpClient, clientErr := config.DefaultHTTPClient(nil)
		if clientErr != nil {
			return nil, clientErr
		}
		client := splunk.NewClient(
			httpClient,
			cfg.Splunk.Target,
			cfg.Splunk.Token(),
			cfg.Splunk.Source,
			cfg.Splunk.SourceType,
			cfg.Splunk.Index,
		)
		var levels []logrus.Level
		for _, l := range logrus.AllLevels {
			if l <= cfg.Level {
				levels = append(levels, l)
			}
		}
		hook := splunk.NewHook(client, levels)
		logger.Hooks.Add(hook)
	}

	return logger, err
}
