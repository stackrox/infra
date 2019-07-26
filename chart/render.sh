#!/bin/sh

if [ ! -d ./infra-server/configs ]; then
    echo 'configs directory missing' 1>&2
    exit 1
fi

helm template ./infra-server --namespace infra --name infra-server \
     --set tag=$(make -C .. tag) \
     --set host=test1.demo.stackrox.com \
     --set ip=34.94.91.159
