.DEFAULT_GOAL=build
.PHONY: build test get run tags clean reset


ensure_vendor:
	mkdir -pv vendor

clean:
	go clean .
	go env

get: clean ensure_vendor
	git clone --depth=1 https://github.com/aws/aws-sdk-go ./vendor/github.com/aws/aws-sdk-go
	rm -rf ./src/main/vendor/github.com/aws/aws-sdk-go/.git
	git clone --depth=1 https://github.com/go-ini/ini ./vendor/github.com/go-ini/ini
	rm -rf ./src/main/vendor/github.com/go-ini/ini/.git
	git clone --depth=1 https://github.com/jmespath/go-jmespath ./vendor/github.com/jmespath/go-jmespath
	rm -rf ./src/main/vendor/github.com/jmespath/go-jmespath/.git
	git clone --depth=1 https://github.com/Sirupsen/logrus ./vendor/github.com/Sirupsen/logrus
	rm -rf ./src/main/vendor/github.com/Sirupsen/logrus/.git
	git clone --depth=1 https://github.com/voxelbrain/goptions ./vendor/github.com/voxelbrain/goptions
	rm -rf ./src/main/vendor/github.com/voxelbrain/goptions/.git
	git clone --depth=1 https://github.com/mjibson/esc ./vendor/github.com/mjibson/esc
	rm -rf ./src/main/vendor/github.com/mjibson/esc/.git
	git clone --depth=1 https://github.com/crewjam/go-cloudformation ./vendor/github.com/crewjam/go-cloudformation
	rm -rf ./src/main/vendor/github.com/crewjam/go-cloudformation/.git

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
