#!/bin/sh

ROOT_PATH=$PWD/..

cd $ROOT_PATH/lib/proto/game
protoc  ./*.proto --go_out=plugins=grpc:./
cd $ROOT_PATH/lib/proto/snowflake
protoc  ./*.proto --go_out=plugins=grpc:./
cd $ROOT_PATH/lib/proto/Chat
protoc  ./*.proto --go_out=plugins=grpc:./
cd $ROOT_PATH/lib/proto/Auth
protoc  ./*.proto --go_out=plugins=grpc:./