/*
Package jsontime gives simple marshalling directly to and from time objects

This is a simple package that provides wrapper types for time.Time and time.Duration
that can be marshalled to and from json and yaml.

time.Time already can marshal to and from json, but not yaml. Whereas
time.Duration cannot marshal to and from either. These wrapper types fill in
this gap.

This is extremely helpful for configuration and REST apis communicating via
json. For example

Unmarshalling yaml config:

	type Config struct {
		StartTime      jsontime.StartTime `yaml:"startTime"`
		RotationPeriod jsontime.Duration  `yaml:"retryPeriod"`
	}
	var config Config
	if err := yaml.Unmarshal(responseBodybody, &config); err != nil {
		return err
	}

	// Get the actual time objects, not the jsontime type aliases
	startTime := target.StartTime.Time()
	rotationPeriod := target.RotationPeriod.Duration()

Unmarshalling a response:

	type ResponseType struct {
		RetryPeriod jsontime.Duration  `json:"retryPeriod"`
	}
	var target ResponseType
	if err := json.Unmarshal(responseBody, &target); err != nil {
		// Error captures any parse error on the retryPeriod value
		// rather than having to do that error check yourself
		return err
	}

	retryPeriod := target.RetryPeriod.Duration()

Marshalling a json request/response:

	type RequestBody struct {
		RefreshRate jsontime.Duration `json:"refreshRate"`
	}
	requestBody := RequestBody{
		RefreshRate: jsontime.Duration(2 * time.Minute),
	}
	marshalled, _ := json.Marshal(requestBody)
	resp, err := http.Do(url, method, bytes.NewBuffer(marshalled))

*/
package jsontime
