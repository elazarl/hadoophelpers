#!/bin/bash
find . -name \*_test.go -type f|sed 's/[^/]*$//'|sort -u|\
	xargs -I_ bash -c '(echo testing _&&cd _&&go test)||exit 255'
