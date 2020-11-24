
## Copyright (c) 2020 Detlef Stern
##
## This file is part of zettelstore.
##
## Zettelstore is licensed under the latest version of the EUPL (European Union
## Public License). Please see file LICENSE.txt for your rights and obligations
## under this license.

.PHONY: test check validate race run build

PACKAGE := zettelstore.de/z/cmd/zettelstore

GO_LDFLAGS := -X main.buildVersion=$(shell git describe --tags --always --dirty || echo unknown)
GOFLAGS := -ldflags "$(GO_LDFLAGS)" -tags osusergo,netgo

test:
	go test ./...

check:
	go vet ./...
	~/go/bin/golint ./...

validate: test check

race:
	go test -race ./...

build:
	mkdir -p bin
	go build $(GOFLAGS) -o bin/zettelstore $(PACKAGE)

release:
	mkdir -p releases
	GOARCH=amd64 GOOS=linux go build $(GOFLAGS) -o releases/zettelstore $(PACKAGE)
	GOARCH=arm GOARM=6 GOOS=linux go build $(GOFLAGS) -o releases/zettelstore-arm6 $(PACKAGE)
	GOARCH=amd64 GOOS=darwin go build $(GOFLAGS) -o releases/iZettelstore $(PACKAGE)
	GOARCH=amd64 GOOS=windows go build $(GOFLAGS) -o releases/zettelstore.exe $(PACKAGE)

clean:
	rm -rf bin releases
