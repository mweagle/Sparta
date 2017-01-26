.DEFAULT_GOAL=build
.PHONY: build test run tags clean reset

clean:
	go clean .
	go env

reset:
		git reset --hard
		git clean -f -d

generate:
	go generate -x
	@echo "Generate complete: `date`"

validate:
	go run $(GOPATH)/src/github.com/fzipp/gocyclo/gocyclo.go -over 20 *.go
	go run $(GOPATH)/src/github.com/fzipp/gocyclo/gocyclo.go -over 20 ./aws
	go run $(GOPATH)/src/github.com/fzipp/gocyclo/gocyclo.go -over 20 ./docker
	go run $(GOPATH)/src/github.com/fzipp/gocyclo/gocyclo.go -over 20 ./explore
	go tool vet *.go
	go tool vet ./explore
	go tool vet ./aws/
	go tool vet ./docker/

format:
	go fmt .

build: format generate validate
	go build .
	@echo "Build complete"

docs:
	@echo ""
	@echo "Sparta godocs: http://localhost:8090/pkg/github.com/mweagle/Sparta"
	@echo
	godoc -v -http=:8090 -index=true

travis-depends:
	go get -u github.com/tools/godep
	godep restore
	rm -rf $(GOPATH)/src/github.com/mjibson/esc
	git clone --depth=1 https://github.com/mjibson/esc $(GOPATH)/src/github.com/mjibson/esc
	rm -rf $(GOPATH)/src/github.com/fzipp/gocyclo
	git clone --depth=1 https://github.com/fzipp/gocyclo $(GOPATH)/src/github.com/fzipp/gocyclo

travis-ci-test: travis-depends build
	go test -v .
	go test -v ./aws/...

test: build
	go test -v .
	go test -v ./aws/...

run: build
	./sparta

tags:
	gotags -tag-relative=true -R=true -sort=true -f="tags" -fields=+l .

provision: build
	go run ./applications/hello_world.go --level info provision --s3Bucket $(S3_BUCKET)

execute: build
	./sparta execute

describe: build
	rm -rf ./graph.html
	go test -v -run TestDescribe
