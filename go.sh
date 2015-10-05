#!/usr/bin/env bash

# Wrapper for the Go tool that I use when gb doesn't provideme something I
# need, like bulding or testing the sources with the race detector.

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
  SOURCE="$(readlink "$SOURCE")"
  [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

GOPATH="$DIR/vendor:$DIR/"
GO=`which go`

env GOPATH=$GOPATH $GO "$@"
