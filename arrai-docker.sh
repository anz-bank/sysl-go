#!/bin/bash

$@ | tar xf - -C /out/
go get -u github.com/rickb777/date
goimports -w /out/

