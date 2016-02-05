#!/bin/bash
echo "Building golinks for ARM..."
GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build 
