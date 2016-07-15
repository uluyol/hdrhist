#!/usr/bin/env bash

set -e

cd "${0%/*}"

scala -classpath lib/*.jar $(find . -name '*.scala') "$@"
