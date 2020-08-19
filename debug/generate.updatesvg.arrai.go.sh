#!/usr/bin/env sh

set -e

OUT=$PWD/updatesvg.arrai.go
cd ../../sysl/pkg/arrai && \
  arrai run concat_go.arrai svg_update.arrai | \
  arrai eval '$"package debug\n\nconst updateScript = `\n${//os.stdin}`"' > $OUT
