#!/bin/sh
docker rm -f snowflake1
docker rmi snowflake
docker build --no-cache --rm=true -t snowflake .
docker run --rm=true --hostname=snowflake --name snowflake1 -it -p 50003:50003 --dns 172.17.0.1 --dns-search service.consul -e LOG_LEVEL=5 snowflake