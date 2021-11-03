#!/bin/bash -xe

export GOARCH="amd64"
export GOOS="darwin" 
go build  -o "build/ocidrive-${GOOS}-${GOARCH}"
export GOOS="linux" 
go build  -o "build/ocidrive-${GOOS}-${GOARCH}"
export GOOS="windows" 
go build  -o "build/ocidrive-${GOOS}-${GOARCH}.exe"
 

