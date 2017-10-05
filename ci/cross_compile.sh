#!/bin/bash
set -ex

VERSION=$1
make clean
make VERSION="$VERSION" packages
make VERSION="$VERSION" GOOS="darwin" TARGET="x86_64-apple-darwin" packages
make VERSION="$VERSION" GOOS="windows" TARGET="x86_64-pc-windows" GO_EXTENSION=".exe" packages
