#!/usr/bin/env bash

set -euo pipefail

IFS=$'\n\t'

error:report() {
    printf '[%s] Error on script: %s, line: %s, return code: %s\n' "${FUNCNAME[0]}" "$0" "$1" "$2"
}
trap 'error:report $LINENO $?' ERR

if (( ${DEBUG:-0} )); then
    set -x
fi

main () {
    mise check
    git push origin master
    git push github master:v2
}

main
