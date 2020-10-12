package authrules

import (
	"context"
)

// Rule is an authorization rule that is is responsible for deciding
// if access to a resource should be allowed or denied.
// If a Rule returns a nil error, this indicates access is allowed.
// If a Rule returns a non-nil error, this indicates that either
// access is denied or some other error was encountered during
// Rule evaluation. If the Rule returns a nil error, it must return
// a non-nil Context. The returned Context may be the input
// ctx or a new Context derived from the input ctx, capturing
// additional values.
type Rule func(ctx context.Context) (context.Context, error)
