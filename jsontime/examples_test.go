package jsontime_test

import (
	"encoding/json"
	"fmt"

	"github.com/anz-bank/sysl-go/jsontime"
	"gopkg.in/yaml.v2"
)

func ExampleDuration_json() {
	type MyStruct struct {
		// Use the yaml tag for marshalling from yaml
		Duration *jsontime.Duration `mapstructure:"duration"`
	}

	const raw = `{"duration": "10m"}`

	var m MyStruct
	_ = json.Unmarshal([]byte(raw), &m)

	fmt.Println(m.Duration)

	// Output: 10m0s
}

func ExampleDuration_yaml() {
	type MyStruct struct {
		// Use the yaml tag for marshalling from yaml
		Duration *jsontime.Duration `mapstructure:"duration"`
	}

	const raw = "duration: 10m"

	var m MyStruct
	_ = yaml.Unmarshal([]byte(raw), &m)

	fmt.Println(m.Duration)

	// Output: 10m0s
}
