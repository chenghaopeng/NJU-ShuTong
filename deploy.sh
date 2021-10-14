#!/bin/sh
set -e

DIR="$( cd "$( dirname "$0"  )" && pwd  )"
cd $DIR

. .env

image=nju-shutong

echo "Compiling"
docker run --rm -v ~/.go/pkg:/go/pkg -v $DIR:/go/src/app golang bash -c "cd /go/src/app && go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=\"-w -s\""

echo "Building"
docker build -t $image .

echo "Pushing"
docker login -u $REGISTRY_USERNAME -p $REGISTRY_PASSWORD registry.cn-hangzhou.aliyuncs.com
docker tag "$image" "registry.cn-hangzhou.aliyuncs.com/chaop-public/$image"
docker push "registry.cn-hangzhou.aliyuncs.com/chaop-public/$image"

echo "Finished!"
