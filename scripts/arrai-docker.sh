#!/bin/bash

$@ | tar xf - -C /out/
goimports -w /out/

