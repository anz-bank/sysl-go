#!/bin/bash

out="$1"; shift

$@ | tar xf - -C "$out"
goimports -w "$out"
