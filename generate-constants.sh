#!/bin/bash -ex

# Create the embedded version
#rm -rf ./resources/provision/node_modules
go run $GOPATH/src/github.com/mjibson/esc/main.go \
  -o ./CONSTANTS.go \
  -private \
  -pkg sparta \
  ./resources

# Create a secondary CONSTANTS_AWSBINARY.go file with empty content.
# The next step will insert the
# build tags at the head of each file so that they are mutually exclusive
go run $GOPATH/src/github.com/mjibson/esc/main.go \
  -o ./CONSTANTS_AWSBINARY.go \
  -private \
  -pkg sparta \
  ./resources/awsbinary/README.md

# Tag the builds...
go run ./cmd/insertTags/main.go ./CONSTANTS !lambdabinary
go run ./cmd/insertTags/main.go ./CONSTANTS_AWSBINARY lambdabinary
