#!/bin/bash

function parentDir() {
  local this_dir=`pwd`
  local child_dir="$1"
  dirname "$child_dir"
  cd $this_dir
}

CURRENT_PATH=$(cd `dirname $0`; pwd)
ROOT_DIR=`parentDir "$CURRENT_PATH"`

VERSION=0.1.0

DIST_DIR=${ROOT_DIR}/dist
echo "Dist directory is ${DIST_DIR}"
rm -rf ${DIST_DIR}
mkdir -p ${DIST_DIR}

echo "Build binary file..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${DIST_DIR}/kuko_linux_amd64_${VERSION} main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${DIST_DIR}/kuko_darwin_amd64_${VERSION} main.go

#cp $ROOT_DIR/docker/Dockerfile $DIST_DIR
#cp $ROOT_DIR/docker/config.yaml $DIST_DIR/config.conf

ls $DIST_DIR