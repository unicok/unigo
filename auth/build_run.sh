#!/bin/sh
docker rm -f chat1
docker build --no-cache --rm=true -t auth .
docker run --rm=true --name auth -it -p 50006:50006 -e SERVICE_NAME=auth -e SERVICE_ID=auth1 auth