#!/bin/sh
docker rm -f game1
docker build --no-cache --rm=true -t game .
docker run --rm=true --hostname=game-dev --name game1 -it -p 51000:51000 -e ETCD_HOST=http://172.17.0.1:4001 -e SERVICE_ID=game1 game