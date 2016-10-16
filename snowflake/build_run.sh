#!/bin/sh
docker rm -f snowflake1
docker build --no-cache --rm=true -t snowflake .
docker run --rm=true --hostname=snowflake --name snowflake1 --network unigo_default -it -p 50003:50003 --dns 172.18.0.1 --dns-search service.consul -e CONSUL_HOST=172.18.0.1 -e CONSUL_HTTP_ADDR=172.18.0.4:8500 -e LOG_LEVEL=5 snowflake