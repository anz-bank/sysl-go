package authexpr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle"
)

var (
	exprParser *participle.Parser
)

//nolint:gochecknoinits // init a single parser as a global to avoid rebuilding it.
func init() {
	// n.b. we need to use a lookahead of 2+ to distinguish between OpExpr and AtomExpr
	exprParser = participle.MustBuild(&Expr{}, participle.UseLookahead(2))
}

type Expr struct {
	OpExpr   *OpExpr `parser:"  @@"`
	AtomExpr *Atom   `parser:"| @@"`
}

// OpExpr look like function calls taking 1 or more Exp arguments.
// We require at least 1 argument so we can distinguish between
// an OpExpr and an Atom. Saying all() or any() is logically
// well-defined but makes it harder to parse input, and isn't very useful.
type OpExpr struct {
	Name string  `parser:"@Ident"`
	Args []*Expr `parser:"\"(\" @@ (\",\" @@)*  \",\"? \")\""`
}

// Atoms look like function calls taking 0 or more Literal arguments.
// This makes them a bit hard to disambiguate from OpExpr. This might
// not be the cleanest or most general way to define a grammar but for
// now it lets the grammar do a bit of type checking for us.
type Atom struct {
	Name string     `parser:"@Ident"`
	Args []*Literal `parser:"\"(\" (@@ (\",\" @@)* )? \",\"? \")\""`
}

type Literal struct {
	String *string `parser:"@String"`
}

func (e *Expr) Validate() error {
	switch {
	case e.OpExpr != nil && e.AtomExpr == nil:
		return e.OpExpr.Validate()
	case e.OpExpr == nil && e.AtomExpr != nil:
		return e.AtomExpr.Validate()
	default:
		panic("authexpr invariant violated: Expr must be either OpExpr or AtomExpr")
	}
}

func (e *OpExpr) Validate() error {
	switch e.Name {
	case "not":
		if len(e.Args) != 1 {
			return ValidationFailed("not(...) OpExpr must be called with exactly one argument")
		}
		fallthrough
	case "all", "any":
	default:
		return ValidationFailed("undefined OpExpr for name: %s", e.Name)
	}
	for _, arg := range e.Args {
		err := arg.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Atom) Validate() error {
	switch e.Name {
	case "jwtHasScope":
		if len(e.Args) != 1 || e.Args[0].String == nil {
			return ValidationFailed("jwtHasScope(...) Atom must be called with exactly one string literal argument")
		}
	default:
		return ValidationFailed("undefined Atom for name: %s", e.Name)
	}
	for _, arg := range e.Args {
		err := arg.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Literal) Validate() error {
	return nil
}

type EvaluationContext struct {
	JWTHasScope func(scope string) (bool, error)
}

func (e *Expr) Evaluate(evalCtx EvaluationContext) (bool, error) {
	if e.OpExpr != nil {
		return e.OpExpr.Evaluate(evalCtx)
	}
	return e.AtomExpr.Evaluate(evalCtx)
}

func (e *OpExpr) Evaluate(evalCtx EvaluationContext) (bool, error) {
	values := make([]bool, len(e.Args))
	for i, arg := range e.Args {
		value, err := arg.Evaluate(evalCtx)
		if err != nil {
			return false, err
		}
		values[i] = value
	}
	switch e.Name {
	case "not":
		return !values[0], nil
	case "any":
		for _, value := range values {
			if value {
				return true, nil
			}
		}
		return false, nil
	case "all":
		for _, value := range values {
			if !value {
				return false, nil
			}
		}
		return true, nil
	default:
		return false, ValidationFailed("undefined OpExpr for name: %s", e.Name)
	}
}

func (e *Atom) Evaluate(evalCtx EvaluationContext) (bool, error) {
	switch e.Name {
	case "jwtHasScope":
		return evalCtx.JWTHasScope(*(e.Args[0].String))
	default:
		return false, ValidationFailed("undefined Atom for name: %s", e.Name)
	}
}

func CompileExpression(expression string) (*Expr, error) {
	root := &Expr{}
	err := exprParser.ParseString(expression, root)
	if err != nil {
		return nil, ParseFailed("failed to parse auth expression").WithCause(err)
	}
	err = root.Validate()
	if err != nil {
		return nil, err
	}
	return root, nil
}

func (e *Expr) Repr() string {
	switch {
	case e.OpExpr != nil && e.AtomExpr == nil:
		return e.OpExpr.Repr()
	case e.OpExpr == nil && e.AtomExpr != nil:
		return e.AtomExpr.Repr()
	default:
		panic("authexpr invariant violated: Expr must be either OpExpr or AtomExpr")
	}
}

func (e *OpExpr) Repr() string {
	reprArgs := make([]string, len(e.Args))
	for i, arg := range e.Args {
		reprArgs[i] = arg.Repr()
	}
	return fmt.Sprintf("%s(%s)", e.Name, strings.Join(reprArgs, ","))
}

func (e *Atom) Repr() string {
	reprArgs := make([]string, len(e.Args))
	for i, arg := range e.Args {
		reprArgs[i] = arg.Repr()
	}
	return fmt.Sprintf("%s(%s)", e.Name, strings.Join(reprArgs, ","))
}

func (e *Literal) Repr() string {
	return strconv.Quote(*e.String)
}
