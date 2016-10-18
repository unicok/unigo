#!/bin/sh
docker rm -f auth
docker rmi auth
docker build --no-cache --rm=true -t auth .
docker run --rm=true --name auth --hostname auth -it -p 50006:50006 --dns 172.17.0.1 --dns-search service.consul -e SERVICE_NAME=auth auth