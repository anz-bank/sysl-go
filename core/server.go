package core

// MARKED TO IGNORE COVERAGE

import (
	"context"

	"github.com/anz-bank/sysl-go/logconfig"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/common"
	"github.com/sirupsen/logrus"
)

type emptyWriter struct {
}

func (e *emptyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// initialise the logger
// sysl-go always uses a pkg logger internally. if custom code passes in a logrus logger, a
// mechanism which is deprecated, then a hook is added to the internal pkg logger that forwards
// logged events to the provided logrus logger.
// sysl-go can be requested to log in a verbose manner. logger in a verbose manner logs additional
// details within log events where appropriate. the mechanism to set this verbose manner is to
// either have a sufficiently high logrus log level or the verbose mode set against the pkg logger.
func InitialiseLogging(ctx context.Context, configs []log.Config, logrusLogger *logrus.Logger) context.Context {
	verboseLogging := false
	if logrusLogger != nil {
		// set an empty io writter against pkg logger
		// pkg logger just becomes a proxy that forwards all logs to logrus
		configs = append(configs,
			log.SetOutput(&emptyWriter{}),
			log.AddHooks(&logrusHook{logrusLogger}),
			log.SetLogCaller(logrusLogger.ReportCaller),
		)
		ctx = common.LoggerToContext(ctx, logrusLogger, nil)
		verboseLogging = logrusLogger.Level >= logrus.DebugLevel
	}

	ctx = log.WithConfigs(configs...).Onto(ctx)
	verboseMode := log.SetVerboseMode(true)
	for _, config := range configs {
		if config == verboseMode {
			verboseLogging = true
			break
		}
	}

	// prepare the middleware
	return logconfig.SetVerboseLogging(ctx, verboseLogging)
}
