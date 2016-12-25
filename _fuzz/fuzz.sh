#!/bin/sh

set -e

if [ `dirname $0` != "." ]; then
	echo "should run in the _fuzz directory"
	exit 1
fi

go-fuzz-build github.com/hirochachacha/plua
go-fuzz -bin plua-fuzz.zip -workdir .
