#!/bin/sh
docker rm -f game1
docker build --no-cache --rm=true -t game .
docker run --rm=true --hostname=game-dev --name game1 -it -p 51000:51000 --dns 172.18.0.1 --dns-search service.consul -e CONSUL_HOST=172.18.0.1 -e SERVICE_ID=game1 game