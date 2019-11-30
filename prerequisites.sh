 #!/bin/bash -ex

GO_GET_FLAGS="-u -v"

# Workaround for https://github.com/golang/go/issues/30515
mkdir -pv ./.sparta
cd ./.sparta   
# Prerequisites
GO111MODULE=off go get $GO_GET_FLAGS github.com/magefile/mage
GO111MODULE=off go get $GO_GET_FLAGS github.com/hhatto/gocloc
GO111MODULE=off go get $GO_GET_FLAGS github.com/mholt/archiver
GO111MODULE=off go get $GO_GET_FLAGS github.com/pkg/browser
GO111MODULE=off go get $GO_GET_FLAGS github.com/otiai10/copy
GO111MODULE=off go get $GO_GET_FLAGS github.com/pkg/errors
GO111MODULE=off go get $GO_GET_FLAGS honnef.co/go/tools/cmd/...
GO111MODULE=off go get $GO_GET_FLAGS github.com/atombender/go-jsonschema/...

# Static analysis
GO111MODULE=off go get $GO_GET_FLAGS honnef.co/go/tools/cmd/...
GO111MODULE=off go get $GO_GET_FLAGS golang.org/x/tools/cmd/goimports
GO111MODULE=off go get $GO_GET_FLAGS github.com/fzipp/gocyclo
GO111MODULE=off go get $GO_GET_FLAGS golang.org/x/lint/golint
GO111MODULE=off go get $GO_GET_FLAGS github.com/mjibson/esc
GO111MODULE=off go get $GO_GET_FLAGS github.com/securego/gosec/cmd/gosec
GO111MODULE=off go get $GO_GET_FLAGS github.com/alexkohler/prealloc
GO111MODULE=off go get $GO_GET_FLAGS github.com/client9/misspell/cmd/misspell
