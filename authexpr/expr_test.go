package authexpr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func demoScopes(scopes []string) func(string) (bool, error) {
	return func(queryScope string) (bool, error) {
		for _, scope := range scopes {
			if queryScope == scope {
				return true, nil
			}
		}
		return false, nil
	}
}

func TestRepr(t *testing.T) {
	t.Parallel()

	type scenario struct {
		input          string
		expectedOutput string
	}

	scenarios := []scenario{
		{
			input:          `all(jwtHasScope("test"))`,
			expectedOutput: `all(jwtHasScope("test"))`,
		},
		{
			input:          `any(jwtHasScope("test"))`,
			expectedOutput: `any(jwtHasScope("test"))`,
		},
		{
			input:          `not(jwtHasScope("test"))`,
			expectedOutput: `not(jwtHasScope("test"))`,
		},
		{
			input:          `not(jwtHasScope('test'))`,
			expectedOutput: `not(jwtHasScope("test"))`,
		},
		{
			input:          `not(jwtHasScope("te'''\"st"))`,
			expectedOutput: `not(jwtHasScope("te'''\"st"))`,
		},
		{
			input:          `all(any(jwtHasScope("fizz"),jwtHasScope("buzz")),not(jwtHasScope("test")))`,
			expectedOutput: `all(any(jwtHasScope("fizz"),jwtHasScope("buzz")),not(jwtHasScope("test")))`,
		},
		{
			input:          `all(any(jwtHasScope("fizz"),jwtHasScope("buzz")),not(jwtHasScope("test")))`,
			expectedOutput: `all(any(jwtHasScope("fizz"),jwtHasScope("buzz")),not(jwtHasScope("test")))`,
		},
		{
			input:          `all(any(jwtHasScope("fizz",),jwtHasScope("buzz",),),not(jwtHasScope("test",),),)`,
			expectedOutput: `all(any(jwtHasScope("fizz"),jwtHasScope("buzz")),not(jwtHasScope("test")))`,
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario // force capture
		t.Run(scenario.input, func(t *testing.T) {
			t.Parallel()
			expr, err := CompileExpression(scenario.input)
			require.NoError(t, err)
			require.Equal(t, scenario.expectedOutput, expr.Repr())
		})
	}
}

func TestOpExprTrailingCommaInvariance(t *testing.T) {
	t.Parallel()

	a, err := CompileExpression(`any(jwtHasScope("foo"), jwtHasScope("barr"))`)
	require.NoError(t, err)

	b, err := CompileExpression(`any(jwtHasScope("foo", ), jwtHasScope("barr", ), )`)
	require.NoError(t, err)

	require.Equal(t, a.Repr(), b.Repr())
}

func TestCompileExpression(t *testing.T) {
	t.Parallel()

	type scenario struct {
		name            string
		inputExprString string
		inputScopes     []string
		expectedResult  bool
		expectedError   string
	}

	scenarios := []scenario{
		{
			name:            "access denied if rule requires scope but there are no scopes",
			inputExprString: `all(jwtHasScope("foo"))`,
			inputScopes:     []string{},
			expectedResult:  false,
			expectedError:   "",
		},
		{
			name:            "access denied if rule requires scope but there is different scope",
			inputExprString: `all(jwtHasScope("foo"))`,
			inputScopes:     []string{"banana"},
			expectedResult:  false,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires scope and there is that scope",
			inputExprString: `all(jwtHasScope("foo"))`,
			inputScopes:     []string{"foo"},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires scope and there is that scope as well as some other scope",
			inputExprString: `all(jwtHasScope("foo"))`,
			inputScopes:     []string{"foo", "banana"},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access denied if rule requires absence of scope but there is that scope",
			inputExprString: `not(jwtHasScope("test"))`,
			inputScopes:     []string{"test", "foo"},
			expectedResult:  false,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires disjunction of scopes and there is one of those scope",
			inputExprString: `any(jwtHasScope("foo"), jwtHasScope("barr"))`,
			inputScopes:     []string{"foo"},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires disjunction of scopes and there is the other of those scope",
			inputExprString: `any(jwtHasScope("foo"), jwtHasScope("barr"))`,
			inputScopes:     []string{"barr"},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires disjunction of scopes and there are both scopes",
			inputExprString: `any(jwtHasScope("foo"), jwtHasScope("barr"))`,
			inputScopes:     []string{"barr", "foo"},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires conjunction of scopes and there are both scopes",
			inputExprString: `all(jwtHasScope("foo"), jwtHasScope("barr"))`,
			inputScopes:     []string{"barr", "foo"},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access denied if rule requires conjunction of scopes and there is only one scope",
			inputExprString: `all(jwtHasScope("foo"), jwtHasScope("barr"))`,
			inputScopes:     []string{"foo"},
			expectedResult:  false,
			expectedError:   "",
		},
		{
			name:            "access denied if rule requires conjunction of scopes and there is only the other scope",
			inputExprString: `all(jwtHasScope("foo"), jwtHasScope("barr"))`,
			inputScopes:     []string{"barr"},
			expectedResult:  false,
			expectedError:   "",
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario // force capture
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			expr, err := CompileExpression(scenario.inputExprString)
			require.NoError(t, err)
			evalCtx := EvaluationContext{
				JWTHasScope: demoScopes(scenario.inputScopes),
			}
			actualResult, err := expr.Evaluate(evalCtx)
			if scenario.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, scenario.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, scenario.expectedResult, actualResult)
			}
		})
	}
}
