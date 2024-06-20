# Based on the Example from Joel Homes, author of "Shipping Go" at
# https://github.com/holmes89/hello-api/blob/main/ch10/Makefile

SHELL=/bin/bash

GO_VERSION := 1.22  # <1>

COVERAGE_AMT := 60  # should be 80

HEREGOPATH := `go env GOPATH`

# setup: # <2>
# 	install-go
# 	init-go
# 
# install-go: # <3>
# 	wget "https://golang.org/dl/go$(GO_VERSION).linux-amd64.tar.gz"
# 	sudo tar -C /usr/local -xzf go$(GO_VERSION).linux-amd64.tar.gz
# 	rm go$(GO_VERSION).linux-amd64.tar.gz
# 
# init-go: # <4>
#     echo 'export PATH=$$PATH:/usr/local/go/bin' >> $${HOME}/.bashrc
#     echo 'export PATH=$$PATH:$${HOME}/go/bin' >> $${HOME}/.bashrc
# 
# upgrade-go: # <5>
# 	sudo rm -rf /usr/bin/go
# 	wget "https://golang.org/dl/go$(GO_VERSION).linux-amd64.tar.gz"
# 	sudo tar -C /usr/local -xzf go$(GO_VERSION).linux-amd64.tar.gz
# 	rm go$(GO_VERSION).linux-amd64.tar.gz

update-javascript:
	# htmx (as at 20 June 2024)
	wget -o web/static/htmx.min.js https://unpkg.com/browse/htmx.org@2.0.0/dist/htmx.min.js
	# hyperscript (as at 20 June 2024)
	wget -o web/static/hyperscript.min.js https://unpkg.com/hyperscript.org@0.9.12/dist/_hyperscript.min.js

build:
	go test ./... && echo "---ok---" && go build -o timeaway cmd/main.go

build-dev:
	go test ./... && echo "---ok---" && go build -o timeaway -tags=development cmd/main.go

test:
	go test ./... -coverprofile=coverage.out

coverage-verbose:
	go tool cover -func coverage.out | tee cover.rpt

coverage-ok:
	cat cover.rpt | grep "total:" | awk '{print ((int($$3) > ${COVERAGE_AMT}) != 1) }'

cover-report:
	go tool cover -html=coverage.out -o cover.html

clean:
	rm cover.html coverage.out cover.rpt

check: check-format check-vet test coverage-verbose coverage-ok cover-report lint 

check-format: 
	test -z $$(go fmt ./...)

check-vet: 
	test -z $$(go vet ./...)

testme:
	echo $(HEREGOPATH)

install-lint:
	# https://golangci-lint.run/usage/install/#local-installation to GOPATH
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(HEREGOPATH)/bin v1.54.2
	# report version
	golangci-lint --version

lint:
	# golangci-lint run -v ./... 
	golangci-lint run ./... 

module-update-tidy:
	go get -u ./...
	go mod tidy

