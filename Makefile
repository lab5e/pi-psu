ifeq ($(GOPATH),)
GOPATH := $(HOME)/go
endif

all: test lint vet build

clean:
	@rm -f bin/*
	@rm -rf pkg/apipb openapiv2
	@go clean -testcache

build: pi-psu flip

pi-psu:
	@echo "*** building $@"
	@cd cmd/$@ && go build -trimpath -o ../../bin/$@ -tags osusergo,netgo

flip:
	@echo "*** building $@"
	@cd cmd/$@ && go build -trimpath -o ../../bin/$@ -tags osusergo,netgo

test:
	@echo "*** test"
	@go test ./...

test_clean:
	@go clean -testcache 
	@go test ./... -timeout 30s

test_verbose:
	@go test ./... -v

test_race:
	@go test ./... -race

lint:
	@echo "*** linting"
	@revive ./... 

vet:
	@echo "*** vetting"
	@go vet ./...

install:
	@sudo systemctl stop pi-psu.service && sudo cp bin/pi-psu /usr/local/bin/. && sudo systemctl start pi-psu.service

staticcheck:
	@staticcheck ./...

benchmark:
	@go test -bench . ./...

count:
	@echo "Linecounts excluding generated and third party code"
	@gocloc --not-match-d='apipb|openapiv2' .

check:
	@echo "*** number of grpc-gateway version fuckups:"
	@find . -name \*.go | xargs grep "grpc-gateway" | grep -v v2 | grep -v "//" | wc -l

dep-install:
	go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.5.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.5.0
	go get github.com/bufbuild/buf/cmd/buf@v0.51.1
	go get github.com/mgechev/revive@latest
