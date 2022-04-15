#! /bin/bash

BASE_PATH=$(pwd)

echo "$BASE_PATH/apiserver"
cd $BASE_PATH/apiserver
go build 

echo "$BASE_PATH/cronserver"
cd $BASE_PATH/cronserver
go build 

echo "$BASE_PATH/mqcserver"
cd $BASE_PATH/mqcserver
go build 

echo "$BASE_PATH/rpcserver"
cd $BASE_PATH/rpcserver
go build 

