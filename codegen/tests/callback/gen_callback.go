package callback

import (
	"context"
)

// GenCallback callbacks used by the generated code
type GenCallback interface {
	DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc)
}
