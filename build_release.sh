#!/bin/bash

if [ -z "$1" ]; then
    echo "Please provide a version number, like 1.0.0.";
    exit;
else
    echo "Building release version '$1'...";
fi

mkdir -p pirsch
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" cmd/main.go
CGO_ENABLED=0 GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" cmd/main.go
mv main pirsch/pirschproxy
mv main.exe pirsch/pirschproxy.exe
cp config.toml pirsch
cp README.md pirsch
cp CHANGELOG.md pirsch
cp LICENSE pirsch

zip -r "pirsch_proxy_v$1.zip" pirsch
rm -r pirsch

echo "Done!"
