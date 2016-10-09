#!/bin/sh
docker rm -f snowflake1
docker build --no-cache --rm=true -t snowflake .
docker run --rm=true --hostname=snowflake-dev --name snowflake1 -it -p 50003:50003 -e ETCD_HOST=http://172.17.0.1:4001 snowflake