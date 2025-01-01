#!/bin/bash

# # Zielplattform definieren (Standard: Linux ARM64)
# GOOS=${GOOS:-linux}   # Betriebssystem
# GOARCH=${GOARCH:-arm64} # Architektur (arm64 für 64-Bit, arm für 32-Bit)
# GOARM=${GOARM:-7}      # ARM-Version (nur für 32-Bit ARM relevant)

# # Git-Commit-Hash für Debugging
# COMMIT=$(git rev-parse HEAD)

# # Statisches Go-Binary bauen
# CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH GOARM=$GOARM go build -ldflags="-s -w -X main.Commit=$COMMIT"#!/bin/bash

CGO_ENABLED=0 go build -ldflags="-s -w -X main.Commit=$(git rev-parse HEAD)"
