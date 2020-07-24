#!/bin/bash

$@ | tar xf - -C /out/
go get -u github.com/rickb777/date # sysl-go and goimports doesn't seem to import this one; TODO: investigate this
goimports -w /out/

