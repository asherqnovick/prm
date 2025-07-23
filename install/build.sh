#!/bin/bash
# build.sh

GOOS=linux GOARCH=amd64 go build -o ../bin/prm_linux_amd64 ../src/main.go
GOOS=darwin GOARCH=amd64 go build -o ../bin/prm_darwin_amd64 ../src/main.go
GOOS=darwin GOARCH=arm64 go build -o ../bin/prm_darwin_arm64 ../src/main.go

lipo -create -output ../bin/prm_darwin_universal ../bin/prm_darwin_amd64 ../bin/prm_darwin_arm64
rm ../bin/prm_darwin_amd64
rm ../bin/prm_darwin_arm64
echo "build complete"
