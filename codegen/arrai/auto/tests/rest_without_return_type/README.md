# Purpose

Simple test of codegen for arrai/auto approach, focusing on Sysl specs whose return statement has no type.

For examples:

``` sysl
/ping1/{identifier <: int}:
    GET:
        return ok
        return ok <: Pong

/ping2/{identifier <: int}:
    GET:
        return ok
        return ok
        return ok <: Pong  
```
