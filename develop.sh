#!/bin/sh
set -e

DIR="$( cd "$( dirname "$0"  )" && pwd  )"
cd $DIR

docker run --rm -it \
  --name nju-shutong \
  -v ~/.go/pkg:/go/pkg \
  -v $DIR:/go/src/app \
  -v $DIR/scripts:/scripts \
  -w /go/src/app \
  golang \
  sh -c ". ./.env && go run ./main.go"
