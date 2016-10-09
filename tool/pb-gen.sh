#!/bin/sh

cd ..;cd lib/proto
protoc  ./*.proto --go_out=plugins=grpc:.