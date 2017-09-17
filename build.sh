#!/bin/sh

GOOS=linux go build

docker build -t vicanso/wechat-enterprise .

rm ./wechat-enterprise
