#!/bin/sh
docker rm -f chat1
docker build --no-cache --rm=true -t auth .
docker run --rm=true --hostname=auth-dev --name auth1 -it -p 50006:50006 -e ETCD_HOST=http://172.17.0.1:4001 -e SERVICE_NAME=auth -e SERVICE_ID=auth1 auth