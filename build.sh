#!/bin/bash

CGO_ENABLED=0 go build -ldflags="-s -w -X main.Commit=$(git rev-parse HEAD)"