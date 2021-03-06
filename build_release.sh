#!/bin/bash

if [ -z "$1" ]; then
    echo "Please provide a version number, like 1.0.0.";
    exit;
else
    echo "Building release version '$1'...";
fi

mkdir -p pirsch
cd js && npm i && npm run build && cd ..
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" main.go
CGO_ENABLED=0 GOOS=windows go build -a -installsuffix cgo -ldflags "-s -w" main.go
mv main pirsch/pirsch-proxy
mv main.exe pirsch/pirsch-proxy.exe
cp config.toml pirsch

zip -r "pirsch_proxy_v$1.zip" pirsch
rm -r pirsch

echo "Done!"
