#!/bin/sh

set -e

while getopts ":t:" opt; do
    case "$opt" in
        t)
            TEMPLATE=$OPTARG
            ;;
        \?)
            echo "Invalid option: -$OPTARG"
            exit 1
            ;;
    esac
done

shift `expr $OPTIND - 1` || :
SYSL_FILE="$1"
if [ ! -z "$SYSL_FILE" ]; then
    shift
fi
GO_MOD="$1"
if [ ! -z "$GO_MOD" ]; then
    shift
fi
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

    while [ "$moreApps" = "y" ]; do
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
arrai run --out=/work/Makefile /sysl-go/codegen/arrai/auto/makefile.arrai "$TEMPLATE" "$SYSL_FILE" $SYSL_APPS
cp /sysl-go/codegen/arrai/auto/codegen.mk /work/codegen.mk
if [ ! -f "go.mod" ]; then
    go mod init "$GO_MOD"
    codegenVersion="${SYSLGO_VERSION##*/}"
    echo "\nrequire github.com/anz-bank/sysl-go ${codegenVersion} // indirect" >> go.mod
    go mod download github.com/anz-bank/sysl-go
fi
printf "\e[1;32mCodegen ready!\e[0m To generate code, run make.\n"
