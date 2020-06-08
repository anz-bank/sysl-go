package core

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	pkgLog "github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/common"
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
			ctx, hook := common.NewTestContextWithLoggerHook()
			logger := pkgLog.From(ctx)
			re := regexp.MustCompile(`hit`)
			writer := &TLSLogFilter{logger, re}
			serverLogger := log.New(writer, "", 0)

			serverLogger.Printf(tt.in)

			require.Equal(t, 1, len(hook.Entries))
			require.Equal(t, tt.in, hook.LastEntry().Message)
		})
	}
}
