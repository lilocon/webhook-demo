#!/bin/sh

set -ex

source .env

docker build . -t ${IMAGE}