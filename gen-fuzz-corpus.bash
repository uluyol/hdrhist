#!/usr/bin/env bash

set -e

cd "${0%/*}"

trap "rm -f fuzz-fuzz.zip" EXIT SIGINT SIGTERM SIGQUIT

go-fuzz-build github.com/uluyol/hdrhist/internal/fuzz
go-fuzz -bin=./fuzz-fuzz.zip -workdir=testdata/fuzz
exit 0
