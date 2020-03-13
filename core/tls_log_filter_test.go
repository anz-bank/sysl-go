package core

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestTLSLogFilter_Write(t *testing.T) {
	type testData struct {
		in    string
		level logrus.Level
	}

	for i, tt := range []testData{
		{
			in:    "hit hit\n",
			level: logrus.DebugLevel,
		},
		{
			in:    "misssssss\n",
			level: logrus.WarnLevel,
		},
	} {
		tt := tt
		t.Run(fmt.Sprintf("TestTLSLogFilter_Write-%d", i), func(t *testing.T) {
			logger, hook := test.NewNullLogger()
			logger.Level = logrus.TraceLevel

			re := regexp.MustCompile(`hit`)
			writer := &TLSLogFilter{logger, re}
			serverLogger := log.New(writer, "", 0)

			serverLogger.Printf(tt.in)

			require.Equal(t, 1, len(hook.Entries))
			require.Equal(t, tt.in, hook.LastEntry().Message)
			require.Equal(t, tt.level, hook.LastEntry().Level)
		})
	}
}
