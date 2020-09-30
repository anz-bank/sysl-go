# jsontime

## Summary

This package provide simple marshalling to and from json and yaml. This is
something lacking in the std `time` package which forces code to check time
related fields after marshalling.

The two main use cases are in json request/responses for microservice
communications, and in reading configuration files. This package provides a
few aliases for the basic type `time.Time` and `time.Duration` specifically
for use when marshalling to and from json and yaml.

## What go time provide

The std `time` package provides two main time related objects to handle time
in go.

* `time.Time` for absolute times, from year down to nanosecond
* `time.Duration` for lengths of time (eg: a wait period)

Time is able to marshal to and from json, but not yaml. There is even plans to
remove this functionality in go 2.0 (this is a note in the source code).

Time cannot marshal to yaml (as yaml marhsalling is done via a third party
package), and Duration cannot do either.

The inability to marshal duration is felt more in configuration management,
where is is common to configure periodic or retry behaviour via durations.

## What this package does

To solve this problem, this package provides aliases for the two basic types,
`time.Time` and `time.Duration`, and implements the four marshalling functions
for each of them

* `MarshalJSON`
* `UnmarshalJSON`
* `MarshalYAML`
* `UnmarshalYAML`

These implementations heavily lean on parsing and stringify logic for the
underlying type in the `time` package, with these methods only service to
read and produce valid json and yaml.

**NOTE:** Time values must be in valid RFC3339 format. Marshalling also
produces time strings in this format.

## What this package does NOT do

This package does NOT expose the full time package api. The aliased types are
basically useless on their own, and only intended to be used for reading and
writing json and yaml. Having said that, there are two things they can do on
top of marshalling,

1. Get the aliased type
2. Stringify

Finally, this package also does not test yaml marshalling, since that would add
another third party dependency.

## Recommended usage

Over a rest api, `time.Time` already has json marshalling, so it is only useful
if needing to read yaml.

Moreover, go duration strings are not standard format when communicating time
durations over apis.

Duration is more intended to be used in config (be it json or yaml). Time can
also be used in config when needed, although we expect this to be seen less
often.

### TL;DR

* Rest APIs: avoid using `Duration`, there is no real difference using `Time`
* Config: Chances are you will use `Duration`, and possibly `Time`.

## Future ideas

It is unclear whether we need to implement marshalling for timestamps. These
are another common method for representing points in time.

It is suspected just integers will do, although the author is unsure how many
API's send timestamps as strings. Handling these would require an `int` alias
with special marshalling logic to handle either strings or ints.

We would also like to solve the yaml marshalling issue, though there are no
ideas on how to do this yet.
