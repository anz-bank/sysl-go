authexpr
========

An authorization expression parsing and evaluation library.

### Purpose

This library allows rules to control authorization to be defined as "authorization expressions" in a simple language.

For examples of syntax, see the unit tests in `expr_test.go`.

### Features

* basic expression parser implemented using https://github.com/alecthomas/participle
* supports evaluating boolean expressions involving `all(...)` `any(...)` `not(...)` and `jwtHasScope("someStringLiteral")`
* can evaluate expression given a decoded JSON claims object in input
* implementation of `jwtHasScope` is abstracted and may be customised.
* includes an implementation of `jwtHasScope` evaluation using the standard definition of the "scope" claim as defined in https://tools.ietf.org/html/rfc8693
