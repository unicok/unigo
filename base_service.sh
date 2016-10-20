#!/bin/sh

docker-compose -f discover-compose.v1.yml down
docker-compose -f db-compose.v1.yml down
docker-compose -f grafana-compose.v1.yml down

docker-compose -f discover-compose.v1.yml up -d
docker-compose -f db-compose.v1.yml up -d
docker-compose -f grafana-compose.v1.yml up -d

curl http://127.0.0.1:8500/v1/kv/seqs/userid -XPUT -d "0"
curl http://127.0.0.1:8500/v1/kv/seqs/snowflake-uuid -XPUT -d "0"