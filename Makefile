
## Copyright (c) 2020 Detlef Stern
##
## This file is part of zettelstore.
##
## Zettelstore is free software: you can redistribute it and/or modify it under
## the terms of the GNU Affero General Public License as published by the Free
## Software Foundation, either version 3 of the License, or (at your option)
## any later version.
##
## Zettelstore is distributed in the hope that it will be useful, but WITHOUT
## ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
## FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
## for more details.
##
## You should have received a copy of the GNU Affero General Public License
## along with Zettelstore. If not, see <http://www.gnu.org/licenses/>.

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
