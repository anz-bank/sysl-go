# jsontime

Solve your time marshalling problems with this barebones `time` package wrapper.

## Usage

This package provides two aliases for the basic `time` objects,

* `type Time time.Time`
* `type Duration time.Duration`

These aliases can be marshalled to and from any json and yaml document. Just
use these types in your fields that you intend to marshal/unmarshal a time
object.

```go
import "github.com/anz-bank/sysl-go/jsontime"

type Response struct {
    ExpiryTime    jsontime.Time     `json:"expiry"`
    RefreshPeriod jsontime.Duration `json:"refresh"`
    // Other fields, make sure they are exported!
}
```

You can the set the response with a single line, instead of stringifying first

```go
var resp := &Response{
    // Set expiry time to now + 5 mins
    ExpiryTime: jsontime.Time(time.Now().Add(5 * time.Minute))

    // Set refresh period to 1 minute
    RefreshPeriod: jsontime.Duration(time.Minute)
}
```

Now you can marshal the response body

```go
// Marshal response body
body, _ := json.Marshal(resp)
```

Or read a response body

```go
// Read response body
var r Response
_ := json.Unmarshal(body, &r)

// Get the underlying types
exp := r.ExpiryTime.Time()
refresh := r.RefreshPeriod.Duration()
```

This nice thing about this last one is that validation of the time and duration
strings are done in the unmarshal function, and the returned type is already in
`time` types.

## Recommendations

Go duration strings are not standard format for REST APIs, so avoid using them
for Request/Response types. There is no real difference if you use this
packages time, or `time.Time`.

Use both freely in config (especially when reading yaml). Chances are you will
use `Duration` much more than `Time` in config though.