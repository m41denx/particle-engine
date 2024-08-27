#!/bin/bash

RED='\033[0;31m'
BLUE='\033[0;34m'
GREEN='\033[0;32m'
NC='\033[0m'
GRAY='\033[0;37m'

#TARGETS="windows-7.0/amd64,windows-7.0/386,linux/amd64,linux/386,linux/arm64,darwin-10.14/amd64,darwin-10.14/arm64"
TARGETS="windows-7.0/amd64,linux/amd64,linux/arm64,darwin-10.14/amd64,darwin-10.14/arm64"

declare -A DISTRIB
DISTRIB["darwin-10.14-amd64"]="d64"
DISTRIB["darwin-10.14-arm64"]="d64a"
DISTRIB["windows-7.0-amd64.exe"]="w64"
#DISTRIB["windows-7.0-386.exe"]="w32"
DISTRIB["linux-amd64"]="l64"
#DISTRIB["linux-386"]="l32"
DISTRIB["linux-arm64"]="l64a"

PATH=$PATH:~/go/bin:
echo -e "${GREEN}Resolving deps...${GRAY}"
go mod tidy
go install github.com/gordonklaus/ineffassign@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install src.techknowlogick.com/xgo@latest
echo -e "${GREEN}Checking for ineffectual assignments...${GRAY}"
ineffassign ./...
echo -e "${GREEN}Checking for cyclomatic complexity${GRAY}"
gocyclo -over 15 .
echo -e "${GREEN}Preparing Build Box...${GRAY}"
docker pull --platform linux/amd64 techknowlogick/xgo:latest
echo -e "${GREEN}Building binaries...${GRAY}"
mkdir -p ./build
for target in $TARGETS; do
  echo -e "${BLUE}Target: ${target}...${GRAY}"
done
rm -rf ./build/*
xgo -targets "$TARGETS" -ldflags="-s -w -X main.BuildTag=$(git rev-parse --short HEAD) -X main.BuildDate=$(date '+%Y-%m-%dT%H:%M')" -dest ./build/ -out particle -pkg cmd .
