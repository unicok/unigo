#!/bin/sh
docker rm -f game1
docker rmi game
docker build --no-cache --rm=true -t game .
docker run --rm=true --hostname game --name game1 -it -p 51000:51000 --dns 172.17.0.1 --dns-search service.consul -e SERVICE_ID=game1 game