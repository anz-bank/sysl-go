package validator

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	vv9 "gopkg.in/go-playground/validator.v9"
)

func TestObjectWithNilPointerFieldTaggedAsNonNilIsInvalid(t *testing.T) {
	// Given an object with pointer field tagged as `validate:"nonnil"``,
	// holding a nil value, when we validate, then we expect the object
	// to be deemed invalid.
	req := require.New(t)

	v := NewDefaultValidator()

	type dummyObj struct {
		OptionalInteger *int `validate:"nonnil"`
	}

	obj := dummyObj{OptionalInteger: nil}

	err := v.Struct(obj)
	req.NotNil(err)
	req.IsType(vv9.ValidationErrors{}, err)
}

func TestObjectWithNonNilPointerFieldTaggedAsNonNilIsValid(t *testing.T) {
	// Given an object with pointer field tagged as `validate:"nonnil"``,
	// holding a non-nil value, when we validate, then we expect the object
	// to be deemed valid.
	req := require.New(t)

	v := NewDefaultValidator()

	type dummyObj struct {
		OptionalInteger *int `validate:"nonnil"`
	}

	x := 0
	obj := dummyObj{OptionalInteger: &x}

	err := v.Struct(obj)
	req.Nil(err)
}

func TestObjectWithNilSliceFieldTaggedAsNonNilIsInvalid(t *testing.T) {
	// Given an object with slice field tagged as `validate:"nonnil"``,
	// holding a nil slice value, when we validate, then we expect the object
	// to be deemed invalid.
	req := require.New(t)

	v := NewDefaultValidator()

	type dummyObj struct {
		Integers []int `validate:"nonnil"`
	}

	// note: Integers initialises to nil slice, not empty slice.
	obj := dummyObj{}

	err := v.Struct(obj)
	req.NotNil(err)
	req.IsType(vv9.ValidationErrors{}, err)
}

func TestObjectWithEmptySliceFieldTaggedAsNonNilIsValid(t *testing.T) {
	// Given an object with slice field tagged as `validate:"nonnil"``,
	// holding a non-nil empty slice value, when we validate, then we expect
	// the object to be deemed valid.
	req := require.New(t)

	v := NewDefaultValidator()

	type dummyObj struct {
		Integers []int `validate:"nonnil"`
	}

	obj := dummyObj{Integers: []int{}}

	err := v.Struct(obj)
	req.Nil(err)
}

func TestValidate(t *testing.T) {
	req := require.New(t)

	req.Error(Validate(1))
}

func TestValidateString(t *testing.T) {
	req := require.New(t)

	req.Nil(Validate("s"))
}

type innerDummyType struct {
	Foo int `validate:"min=5,max=7"`
}

type outerDummyType struct {
	X innerDummyType
}

func TestValidationRecursivelyValidatesStructElements(t *testing.T) {
	// Given:
	//   outer value containing another type as element, where the inner type implements Validator
	// When:
	//	we validate a value of the outer value
	// Then:
	//	contained element validation logic is executed and failed element validation counts
	//  as a failed validation for entire outer value.

	err := Validate(outerDummyType{X: innerDummyType{10}})
	require.Error(t, err)
}

type dummyvalidator struct {
	foo string
}

type dummyTester struct {
	Inner dummyvalidator `validate:"min=3"`
}

var validatorCalled = false

func testDummyValidator(reflect.Value) interface{} {
	validatorCalled = true
	return 6
}

func TestMain(m *testing.M) {
	RegisterCustomValidator(testDummyValidator, dummyvalidator{})
}
func TestRegisterCustomValidator(t *testing.T) {
	// What this test is doing is ensuring that the registered validator is called for the correct type.

	err := Validate(dummyTester{dummyvalidator{"foo"}})
	require.NoError(t, err)
	require.True(t, validatorCalled)

	// And panic if types are registered after init
	require.Panics(t, func() {
		RegisterCustomValidator(func(field reflect.Value) interface{} { return "hello" }, outerDummyType{})
	})
}

type durationValidation struct {
	X time.Duration `validate:"timeout=10ms"`    // max only
	Y time.Duration `validate:"timeout=2ms:10s"` // min,max
	Z time.Duration `validate:"timeout=10ns:"`   // min only
}

func errorDueToFields(t assert.TestingT, err error, field ...string) {
	found := map[string]bool{}
	for _, e := range err.(vv9.ValidationErrors) {
		found[e.StructField()] = true
	}

	if len(found) != len(field) {
		for _, f := range field {
			if _, ok := found[f]; !ok {
				assert.Failf(t, "Field '%s' expected to fail but didn't", f)
			}
		}
	}
}

func TestDurationValidation(t *testing.T) {
	tests := []struct {
		name       string
		in         durationValidation
		errorfield string
	}{
		{
			name:       "pass",
			in:         durationValidation{X: 5 * time.Millisecond, Y: 8 * time.Millisecond, Z: 30 * time.Nanosecond},
			errorfield: "",
		},
		{
			name:       "fail X",
			in:         durationValidation{X: 15 * time.Millisecond, Y: 8 * time.Millisecond, Z: 30 * time.Nanosecond},
			errorfield: "X",
		},
		{
			name:       "fail Y, below min",
			in:         durationValidation{X: 5 * time.Millisecond, Y: 1 * time.Millisecond, Z: 30 * time.Nanosecond},
			errorfield: "Y",
		},
		{
			name:       "fail Y, equal max",
			in:         durationValidation{X: 5 * time.Millisecond, Y: 10 * time.Second, Z: 30 * time.Nanosecond},
			errorfield: "Y",
		},
		{
			name:       "fail Y, above max",
			in:         durationValidation{X: 5 * time.Millisecond, Y: 11 * time.Second, Z: 30 * time.Nanosecond},
			errorfield: "Y",
		},
		{
			name:       "fail Z, below min",
			in:         durationValidation{X: 5 * time.Millisecond, Y: 8 * time.Millisecond, Z: 3 * time.Nanosecond},
			errorfield: "Z",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)
			if tt.errorfield != "" {
				require.Error(t, err)
				errorDueToFields(t, err, tt.errorfield)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
func TestInvalidDurationTags(t *testing.T) {
	err := Validate(struct {
		T time.Duration `validate:"timeout"`
	}{})
	require.Error(t, err)

	err = Validate(struct {
		T time.Duration `validate:"timeout=1xx"`
	}{})
	require.Error(t, err)

	err = Validate(struct {
		T time.Duration `validate:"timeout=1ms:xx"`
	}{})
	require.Error(t, err)

	err = Validate(struct {
		T time.Duration `validate:"timeout=1dd:xx"`
	}{})
	require.Error(t, err)

	err = Validate(struct {
		T time.Duration `validate:"timeout=1dd:xx:"`
	}{})
	require.Error(t, err)

	err = Validate(struct {
		T int `validate:"timeout=1dd:xx:"`
	}{})
	require.Error(t, err)
}
