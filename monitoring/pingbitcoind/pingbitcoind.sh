#!/bin/bash

export GOPATH=/home/api/api
export GOROOT=/home/api/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

source $GOPATH/bin/environment.sh

$GOPATH/bin/pingbitcoind
