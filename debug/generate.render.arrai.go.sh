#!/usr/bin/env sh

set -e

OUT=$PWD/render.arrai.go
cd ../../sysl/pkg/arrai && \
  arrai run concat_go.arrai debug.arrai | \
  arrai eval '$"package debug\n\nconst renderScript = `\n${//os.stdin}`"' > $OUT
