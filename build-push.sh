#!/bin/bash

# prepare
#  echo "aaa" | docker login --username=gha1 --password-stdin container-registry1.chasoba.net

CURRENT_DATE=$(date --utc -u +"%Y%m%d")
GIT_REV_SHORT=$(git rev-parse --short HEAD)
export KO_DOCKER_REPO=container-registry1.chasoba.net/sobadon/lego-patch-selfdns

# --push=false
ko build --bare --tags "${CURRENT_DATE}-${GIT_REV_SHORT}" ./cmd/lego
