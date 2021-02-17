package core

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	anzlog "github.com/anz-bank/sysl-go/log"

	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
)

func TestTLSLogFilter_Write(t *testing.T) {
	type testData struct {
		in string
	}
	for i, tt := range []testData{
		{
			in: "misssssss\n",
		},
	} {
		tt := tt
		t.Run(fmt.Sprintf("TestTLSLogFilter_Write-%d", i), func(t *testing.T) {
			logger := testutil.NewTestLogger()
			re := regexp.MustCompile(`hit`)
			writer := &TLSLogFilter{logger.WithLevel(anzlog.DebugLevel), re}
			serverLogger := log.New(writer, "", 0)

			serverLogger.Printf(tt.in)

			require.Equal(t, 1, logger.EntryCount())
			require.Equal(t, tt.in, logger.LastEntry().Message)
		})
	}
}
