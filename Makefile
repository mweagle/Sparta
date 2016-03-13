.DEFAULT_GOAL=build
.PHONY: build test get run tags clean reset

clean:
	go clean .
	go env

get: clean
	go get -t -u ./...
	go get github.com/mjibson/esc
	go env

reset:
		git reset --hard
		git clean -f -d

generate:
	go generate -x
	@echo "Generate complete: `date`"

format:
	go fmt .

vet: generate
	# Disable composites until https://github.com/golang/go/issues/9171 is resolved.  Currently
	# failing due to gocf.IAMPoliciesList literal initialization
	go tool vet -composites=false *.go
	go tool vet -composites=false ./explore
	go tool vet -composites=false ./aws/

build: format generate vet
	go build .
	@echo "Build complete"

docs:
	@echo ""
	@echo "Sparta godocs: http://localhost:8090/pkg/Sparta/"
	@echo
	godoc -v -http=:8090 -index=true

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
