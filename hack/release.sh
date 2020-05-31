#!/bin/sh

source .env

set -ex

TAG=${1}

git tag ${TAG}

echo $TAG

docker build . -t ${IMAGE}:${TAG}

echo "push"

sleep 3

docker push ${IMAGE}:${TAG}
