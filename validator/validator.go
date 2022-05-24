package validator

import (
	"reflect"
	"strings"
	"time"

	vv10 "github.com/go-playground/validator/v10"
)

// Validator values have a Validate() method returning nil iff the value is
// deemed valid.
type Validator interface {
	Validate() error
}

func NewDefaultValidator() *vv10.Validate {
	v := vv10.New()

	v.RegisterAlias("nonnil", "required")
	_ = v.RegisterValidation("timeout", timeoutValidatorFunc)
	for _, data := range registeredTypes {
		v.RegisterCustomTypeFunc(data.fn, data.types...)
	}
	for _, data := range registeredStructLevel {
		v.RegisterStructValidation(data.fn, data.types...)
	}
	return v
}

var (
	DefaultValidator *vv10.Validate

	registeredTypes       []registrationData
	registeredStructLevel []registrationStructLevelData
)

type registrationData struct {
	fn    vv10.CustomTypeFunc
	types []interface{}
}
type registrationStructLevelData struct {
	fn    vv10.StructLevelFunc
	types []interface{}
}

func RegisterCustomValidator(fn vv10.CustomTypeFunc, types ...interface{}) {
	if DefaultValidator != nil {
		panic("attempting to add a new validator after init()")
	}
	registeredTypes = append(registeredTypes, registrationData{
		fn,
		types,
	})
}

func RegisterStructLevel(fn vv10.StructLevelFunc, types ...interface{}) {
	if DefaultValidator != nil {
		panic("attempting to add a new validator after init()")
	}
	registeredStructLevel = append(registeredStructLevel, registrationStructLevelData{
		fn,
		types,
	})
}

// Validate validates the fields of a struct based
// on 'validator' tags and returns errors found indexed
// by the field name.
func Validate(v interface{}) error {
	if reflect.TypeOf(v).Kind() == reflect.String {
		return nil
	}
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		switch val.Elem().Kind() {
		case reflect.String, reflect.Slice:
			return nil
		}
	}
	if DefaultValidator == nil {
		DefaultValidator = NewDefaultValidator()
	}
	return DefaultValidator.Struct(v)
}

// Var validates a single string using tag style validation.
func ValidateString(val, tag string) error {
	if DefaultValidator == nil {
		DefaultValidator = NewDefaultValidator()
	}
	return DefaultValidator.Var(val, tag)
}

// Custom validator to manage a timeout= param
// timeout=1ms     -> 1ms max timeout, no minimum to validate
// timeout=1ms:10s -> timeout between 1ms (inclusive) and 10s (exclusive)
// timeout=5s:     -> 5s min timeout, no maximum.
func timeoutValidatorFunc(fl vv10.FieldLevel) bool {
	parts := strings.Split(fl.Param(), ":")

	val, ok := fl.Field().Interface().(time.Duration)
	if !ok {
		return false
	}
	switch len(parts) {
	case 1: // max
		p, err := time.ParseDuration(parts[0])
		if err != nil {
			return false
		}
		return val < p
	case 2: // min, max
		min, err := time.ParseDuration(parts[0])
		if err != nil {
			return false
		}
		if len(parts[1]) > 0 {
			max, err := time.ParseDuration(parts[1])
			if err != nil {
				return false
			}
			return val >= min && val < max
		}
		return val >= min
	}
	return false
}
