#!/bin/sh
go build -o release/gosip/gosip
cp -rf config.yml release/gosip/
cd release
tar -zcvf gosip.tar.gz gosip/
cd ..
echo ""
echo "build gosip success"