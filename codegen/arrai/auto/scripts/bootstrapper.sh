#!/bin/sh

set -e

SYSL_FILE="$1"
GO_MOD="$2"
shift 2 || :
SYSL_APPS="$@"

if [ -z "$SYSL_FILE" ]; then
    SYSL_FILE=`find . -type f -iname '*.sysl' | sort -z | head -1`
    read -p "Sysl model path (default: \"$SYSL_FILE\"): " filename
    if [ -n "$filename" ]; then
        SYSL_FILE="$filename"
    fi
fi

if [ -z "$GO_MOD" ]; then
    GO_MOD="$SYSL_FILE"
    GO_MOD="${GO_MOD%.sysl}"
    GO_MOD="${GO_MOD##*/}"
    read -p "Go module name (default: \"$GO_MOD\"): " modname
    if [ -n "$modname" ]; then
        GO_MOD="$modname"
    fi
fi

if [ -z "$SYSL_APPS" ]; then
    index=0
    moreApps="y"

    while [ "$moreApps" == "y" ]; do
        appDefault=`arrai r /sysl-go/codegen/arrai/auto/scripts/get_apps.arrai "$SYSL_FILE" "$index"`
        if [ -z "$appDefault" ]; then
            break
        fi
        pkgDefault=$(echo $appDefault | tr '[:upper:]' '[:lower:]')
        read -p "App to codegen (default: \"$appDefault\"): " app
        if [ -z "$app" ]; then
            app="$appDefault"
        fi
        read -p "Package name for \"$appDefault\" (default: \"$pkgDefault\"): " pkg
        if [ -z "$pkg" ]; then
            pkg="$pkgDefault"
        fi
        SYSL_APPS="$SYSL_APPS ${app}:${pkg}"
        read -p "Add another app (y/N)? " moreApps
        if [ -z "$moreApps" ]; then
            moreApps="y"
        fi
        index=$(($index+1))
    done
fi

cd /work
arrai run --out=/work/Makefile /sysl-go/codegen/arrai/auto/makefile.arrai "$SYSL_FILE" $SYSL_APPS
arrai run --out=/work/codegen.mk /sysl-go/codegen/arrai/auto/codegen.mk.arrai "$SYSLGO_VERSION"
go mod init "$GO_MOD"
printf "\e[1;32mCodegen ready!\e[0m To generate code, run make.\n"
