package common

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dlclark/regexp2"
)

// The builtin regex library only supports RE2 which guarantees to run in constant time, however, is missing features
// that are allowed in regexes specified in an OpenApi spec (Ecma-262 Edition 5.1), for example lookaheads.
// This utility will try to compile with the builtin library first, but if that fails will fall back to the regexp2
// library which supports a lot more features.

type RegexWithFallBack struct {
	stdLibRe *regexp.Regexp
	re2      *regexp2.Regexp
}

func RegexWithFallbackMustCompile(str string) *RegexWithFallBack {
	ret := &RegexWithFallBack{}
	var err error
	ret.stdLibRe, err = regexp.Compile(str)
	if err != nil {
		ret.re2 = regexp2.MustCompile(str, 0)
		// Just set a large timeout which should never be hit
		ret.re2.MatchTimeout = time.Minute
	}

	return ret
}

func (r *RegexWithFallBack) MatchString(s string) bool {
	if r.stdLibRe != nil {
		return r.stdLibRe.MatchString(s)
	}

	ret, err := r.re2.MatchString(s)
	if err != nil {
		// error can only be timeout which can get handled by the standard timeout handler
		panic(fmt.Errorf("regexWithFallBack: %w", err))
	}

	return ret
}
