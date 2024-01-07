#!/bin/bash

RED='\033[0;31m'
BLUE='\033[0;34m'
GREEN='\033[0;32m'
NC='\033[0m'
GRAY='\033[0;37m'

declare -A dists
dists["w64.exe"]="windows/amd64"
dists["w32.exe"]="windows/386"

dists["l64"]="linux/amd64"
dists["l64a"]="linux/arm64"
dists["l32"]="linux/386"

dists["d64"]="darwin/amd64"
dists["d64a"]="darwin/arm64"

PATH=$PATH:~/go/bin
echo -e "${GREEN}Resolving deps...${GRAY}"
go mod tidy
go install github.com/gordonklaus/ineffassign@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
echo -e "${GREEN}Checking for ineffectual assignments...${GRAY}"
ineffassign ./...
echo -e "${GREEN}Building binaries...${GRAY}"
mkdir -p ./build
for arch in "${!dists[@]}"; do
  echo -e "    ${BLUE}Building for ${arch}${NC}"
  GOOS=${dists[$arch]%%/*} GOARCH=${dists[$arch]##*/} go build -ldflags="-s -w -X cmd.BuildTag=$(git rev-parse --short HEAD) -X cmd.BuildDate=$(date '+%Y-%m-%d_%H:%M')" -o ./build/particle-$arch ./cmd/...
done

echo -e "${GREEN}Done.${NC}"