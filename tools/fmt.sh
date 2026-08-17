#!/usr/bin/env bash

echo -e "formating code"

opt=${1-default}
set -euo pipefail

if [ "$opt" = "x" ] || [ "$opt" = "exec" ]; then
    echo -e "writing the formatted output $opt \n"
    gofmt -w .
fi

if [ "$opt" = "l" ] || [ "$opt" = "list" ]; then
    echo -e "listing the source files that need formating $opt \n"
    gofmt -l .
fi

if [ "$opt" = "d" ] || [ "$opt" = "diff" ]; then
    echo -e "show the diff of needed changes and current state $opt \n"
    gofmt -d .
fi
