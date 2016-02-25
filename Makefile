PROJECT=nomi
ORGANIZATION=giantswarm

BUILD_PATH := $(shell pwd)/.gobuild
PROJECT_PATH := $(BUILD_PATH)/src/github.com/$(ORGANIZATION)
IMPORT_PATH := github.com/$(ORGANIZATION)/$(PROJECT)

BIN := $(PROJECT)

VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD)
GOOS := linux
GOARCH := amd64

.PHONY: clean run-test get-deps deps fmt test

GOPATH := $(BUILD_PATH)

SOURCE=$(shell find . -name '*.go')

all: get-deps $(BIN)

ci: clean all test

clean:
		rm -rf $(BUILD_PATH) $(BIN) output/gobindata.go

install: $(BIN)
	cp nomi /usr/local/bin/

get-deps: .gobuild .gobuild/bin/go-bindata

deps:
	@${MAKE} -B -s .gobuild/bin/go-bindata
	@${MAKE} -B -s .gobuild

.gobuild/bin/go-bindata:
	GOPATH=$(GOPATH) GOBIN=$(GOPATH)/bin go get github.com/jteeuwen/go-bindata/...

.gobuild:
	@mkdir -p $(PROJECT_PATH)
	@rm -f $(PROJECT_PATH)/$(PROJECT) && cd "$(PROJECT_PATH)" && ln -s ../../../.. $(PROJECT)
	#
	# Fetch public dependencies via `go get`
	# All of the dependencies are listed here
	@GOPATH=$(GOPATH) go get github.com/jteeuwen/go-bindata/...
	@GOPATH=$(GOPATH) go get github.com/spf13/cobra
	@GOPATH=$(GOPATH) go get github.com/aybabtme/uniplot/histogram
	@GOPATH=$(GOPATH) go get github.com/coreos/fleet/client
	@GOPATH=$(GOPATH) go get github.com/coreos/fleet/schema
	@GOPATH=$(GOPATH) go get github.com/golang/glog
	@GOPATH=$(GOPATH) go get github.com/op/go-logging
	@GOPATH=$(GOPATH) go get github.com/gorilla/mux
	@GOPATH=$(GOPATH) go get gopkg.in/yaml.v2
	@GOPATH=$(GOPATH) go get github.com/ajstarks/svgo
	@GOPATH=$(GOPATH) go get github.com/eapache/queue
	@GOPATH=$(GOPATH) go get github.com/dustin/go-humanize
	@GOPATH=$(GOPATH) go get gopkg.in/check.v1

	# Fetch public dependencies via `go get`
	GOPATH=$(GOPATH) go get -d -v $(IMPORT_PATH)

$(BIN): $(SOURCE) VERSION output/gobindata.go
		echo Building for $(GOOS)/$(GOARCH)
		docker run \
		    --rm \
		    -v $(shell pwd):/usr/code \
		    -e GOPATH=/usr/code/.gobuild \
		    -e GOOS=$(GOOS) \
		    -e GOARCH=$(GOARCH) \
		    -w /usr/code \
		    golang:1.5.3 \
		    go build -a -ldflags "-X $(IMPORT_PATH)/cmd.ProjectVersion=$(VERSION) -X $(IMPORT_PATH)/cmd.ProjectBuild=$(COMMIT)" -o $(BIN)


output/gobindata.go:
		.gobuild/bin/go-bindata -pkg output -o output/gobindata.go ./output/embedded/

test: get-deps
		docker run \
		--rm \
		-v $(shell pwd):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		-e GO15VENDOREXPERIMENT=1 \
		-w /usr/code/ \
	golang:1.5.3 \
		bash -c 'cd .gobuild/src/github.com/$(ORGANIZATION)/$(PROJECT) && go test $$(go list ./... | grep -v "gopath")'

lint:
	GOPATH=$(GOPATH) go vet $(go list ./... | grep -v "gopath")
	GOPATH=$(GOPATH) golint $(go list ./... | grep -v "gopath")

godoc: all
	@echo Opening godoc server at http://localhost:6060/pkg/github.com/$(ORGANIZATION)/$(PROJECT)/
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOROOT=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -e GO15VENDOREXPERIMENT=1 \
	    -w /usr/code \
      -p 6060:6060 \
		golang:1.5 \
		godoc -http=:6060

fmt:
	gofmt -l -w .

bin-dist: all
	mkdir -p bin-dist/
	cp -f README.md bin-dist/
	cp -f LICENSE bin-dist/
	cp $(PROJECT) bin-dist/
	cd bin-dist/ && tar czf $(PROJECT).$(VERSION).tar.gz *
