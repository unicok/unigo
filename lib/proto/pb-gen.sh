#!/bin/sh

ROOT_PATH=$PWD

cd $ROOT_PATH/game
protoc  ./*.proto --go_out=plugins=grpc:./
cd $ROOT_PATH/snowflake
protoc  ./*.proto --go_out=plugins=grpc:./
cd $ROOT_PATH/chat
protoc  ./*.proto --go_out=plugins=grpc:./
cd $ROOT_PATH/auth
protoc  ./*.proto --go_out=plugins=grpc:./