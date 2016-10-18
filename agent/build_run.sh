#!/bin/sh
docker rm -f agent1
docker build --no-cache --rm=true -t agent .
docker run -it --rm=true --hostname=agent --name agent1 -it -p 8888:8888 --dns 172.17.0.1 -e SERVICE_ID=agent1 agent