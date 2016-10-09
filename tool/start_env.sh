#!/bin/sh

docker rm -f etcd
docker rm -f etcd-browser
docker rm -f kamon-grafana-dashboard
docker rm -f registrator


docker run -d --name=etcd -p 4001:4001 -p 7001:7001 -v /Volumes/etcd/:/data microbox/etcd -name=etcd0 -cors=*
docker run -d --name etcd-browser -p 8000:8000 -e ETCD_HOST=127.0.0.1 -e AUTH_PASS=admin buddho/etcd-browser
docker run -d --name kamon-grafana-dashboard -p 80:80 -p 81:81 -p 8125:8125/udp -p 8126:8126 kamon/grafana_graphite
docker run -d --name=registrator --net=host --volume=/var/run/docker.sock:/tmp/docker.sock gliderlabs/registrator -internal etcd://172.17.0.1:4001/backends

curl http://172.17.0.1:4001/v2/keys/seqs/userid -XPUT -d value="0"
curl http://172.17.0.1:4001/v2/keys/seqs/snowflake-uuid -XPUT -d value="0"