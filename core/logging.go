package core

import (
	"fmt"

	"github.com/anz-bank/pkg/log"
	"github.com/sirupsen/logrus"
)

type logrusHook struct {
	logger *logrus.Logger
}

func (h *logrusHook) OnLogged(entry *log.LogEntry) error {
	e := pkgLogEntryToLogrusEntry(h.logger, entry)
	e.Log(e.Level, e.Message)
	return nil
}

// Convert the given pkg entry to a logrus entry.
func pkgLogEntryToLogrusEntry(logger *logrus.Logger, entry *log.LogEntry) *logrus.Entry {
	return &logrus.Entry{
		Logger:  logger,
		Data:    pkgLogEntryToLogrusFields(entry),
		Time:    entry.Time,
		Level:   verboseToLogrusLevel(entry.Verbose),
		Message: entry.Message,
	}
}

// Convert the pkg log entry into appropriate logrus fields to log.
func pkgLogEntryToLogrusFields(entry *log.LogEntry) logrus.Fields {
	fields := make(map[string]interface{})
	iterator := entry.Data.Range()
	for iterator.Next() {
		fields[iterator.Key().(string)] = iterator.Value()
	}
	if entry.Caller.File != "" {
		fields["caller"] = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
	}
	return fields
}

// Convert the pkg concept of verbosity to a logrus level.
func verboseToLogrusLevel(verbose bool) logrus.Level {
	if verbose {
		return logrus.DebugLevel
	}
	return logrus.InfoLevel
}
