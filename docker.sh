#!/usr/bin/env bash
docker run -it -v $GOPATH/src:/go/src -w /go/src/github.com/mintzhao/pjnath-go golang