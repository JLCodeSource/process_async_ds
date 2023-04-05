#!/bin/sh

sudo cp /workspaces/process_async_ds/gbr /usr/bin/gbr
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.52.2
