#!/bin/bash

# Switch to script directory
cd "$(dirname "$0")" || exit

# Create bin directory if it doesn't exist
if [ ! -d "../bin" ]; then
    mkdir -p "../bin"
fi

echo "Building Bfck for Linux..."

# Build the project
cd ../src || exit
if CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ../bin/bfck .; then
    echo -e "\033[0;32mBuild successful! Output: bin/bfck\033[0m"
else
    echo -e "\033[0;31mBuild failed!\033[0m"
    exit 1
fi
