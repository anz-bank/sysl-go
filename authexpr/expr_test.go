package authexpr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testScopeFoo                = "foo"
	testScopeBarr               = "barr"
	exprNotJWTHasScopeTest      = `not(jwtHasScope("test"))`
	exprAllJWTHasScopeFoo       = `all(jwtHasScope("foo"))`
	exprAnyJWTHasScopeFooBarr   = `any(jwtHasScope("foo"), jwtHasScope("barr"))`
	exprAllJWTHasScopeFooBarr   = `all(jwtHasScope("foo"), jwtHasScope("barr"))`
	exprComplexWithFizzBuzzTest = `all(any(jwtHasScope("fizz"),jwtHasScope("buzz")),not(jwtHasScope("test")))`
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
			input:          exprNotJWTHasScopeTest,
			expectedOutput: exprNotJWTHasScopeTest,
		},
		{
			input:          `not(jwtHasScope('test'))`,
			expectedOutput: exprNotJWTHasScopeTest,
		},
		{
			input:          `not(jwtHasScope("te'''\"st"))`,
			expectedOutput: `not(jwtHasScope("te'''\"st"))`,
		},
		{
			input:          exprComplexWithFizzBuzzTest,
			expectedOutput: exprComplexWithFizzBuzzTest,
		},
		{
			input:          exprComplexWithFizzBuzzTest,
			expectedOutput: exprComplexWithFizzBuzzTest,
		},
		{
			input:          `all(any(jwtHasScope("fizz",),jwtHasScope("buzz",),),not(jwtHasScope("test",),),)`,
			expectedOutput: exprComplexWithFizzBuzzTest,
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

	a, err := CompileExpression(exprAnyJWTHasScopeFooBarr)
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
			inputExprString: exprAllJWTHasScopeFoo,
			inputScopes:     []string{},
			expectedResult:  false,
			expectedError:   "",
		},
		{
			name:            "access denied if rule requires scope but there is different scope",
			inputExprString: exprAllJWTHasScopeFoo,
			inputScopes:     []string{"banana"},
			expectedResult:  false,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires scope and there is that scope",
			inputExprString: exprAllJWTHasScopeFoo,
			inputScopes:     []string{testScopeFoo},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires scope and there is that scope as well as some other scope",
			inputExprString: exprAllJWTHasScopeFoo,
			inputScopes:     []string{testScopeFoo, "banana"},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access denied if rule requires absence of scope but there is that scope",
			inputExprString: exprNotJWTHasScopeTest,
			inputScopes:     []string{"test", "foo"},
			expectedResult:  false,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires disjunction of scopes and there is one of those scope",
			inputExprString: exprAnyJWTHasScopeFooBarr,
			inputScopes:     []string{testScopeFoo},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires disjunction of scopes and there is the other of those scope",
			inputExprString: exprAnyJWTHasScopeFooBarr,
			inputScopes:     []string{testScopeBarr},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires disjunction of scopes and there are both scopes",
			inputExprString: exprAnyJWTHasScopeFooBarr,
			inputScopes:     []string{testScopeBarr, testScopeFoo},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access granted if rule requires conjunction of scopes and there are both scopes",
			inputExprString: exprAllJWTHasScopeFooBarr,
			inputScopes:     []string{testScopeBarr, testScopeFoo},
			expectedResult:  true,
			expectedError:   "",
		},
		{
			name:            "access denied if rule requires conjunction of scopes and there is only one scope",
			inputExprString: exprAllJWTHasScopeFooBarr,
			inputScopes:     []string{testScopeFoo},
			expectedResult:  false,
			expectedError:   "",
		},
		{
			name:            "access denied if rule requires conjunction of scopes and there is only the other scope",
			inputExprString: exprAllJWTHasScopeFooBarr,
			inputScopes:     []string{testScopeBarr},
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
