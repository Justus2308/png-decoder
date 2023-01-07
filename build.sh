#!/bin/bash

DIR="$(cd "$(dirname "$0")" && pwd)"
cd $DIR
go build -a ./src/run.go
echo "compiled executable"