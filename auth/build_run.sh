#!/bin/sh
docker rm -f auth
docker rmi auth
docker build --no-cache --rm=true -t auth .
docker run --rm=true --name auth --network unigo_default -it -p 50006:50006 --dns 172.18.0.1 --dns-search service.consul -e DNS_ADDR=172.18.0.1:53 -e SERVICE_NAME=auth -e SERVICE_ID=auth1 auth