#!/bin/sh
docker rm -f chat1
docker build --no-cache --rm=true -t chat .
docker run --rm=true --hostname=chat-dev --name game1 -it -p 50008:50008 -e ETCD_HOST=http://172.17.0.1:4001 -e SERVICE_NAME=chat -e SERVICE_ID=chat1 chat