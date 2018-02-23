all: build

export GOBIN := $(CURDIR)/bin

CURRENT_GIT_GROUP := github.com/luweimy
CURRENT_GIT_REPO  := goproxy-server

deps:

build: deps
	go build -o bin/$(CURRENT_GIT_REPO)  $(CURRENT_GIT_GROUP)/$(CURRENT_GIT_REPO)/

build_linux: deps
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o bin/$(CURRENT_GIT_REPO)-linux64  $(CURRENT_GIT_GROUP)/$(CURRENT_GIT_REPO)/

build_windows: deps
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
	go build -o bin/$(CURRENT_GIT_REPO)-win64  $(CURRENT_GIT_GROUP)/$(CURRENT_GIT_REPO)/

build_macosx: deps
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
	go build -o bin/$(CURRENT_GIT_REPO)-macosx64  $(CURRENT_GIT_GROUP)/$(CURRENT_GIT_REPO)/

test:

clean:
	@rm -rf bin pkg

.PHONY:  deps test clean
