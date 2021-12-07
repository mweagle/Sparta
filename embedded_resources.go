//go:build !lambdabinary

//lint:file-ignore U1000,ST1018 Ignore all unused code, it's generated
/* #nosec */

package sparta

import (
	"embed"
)

//go:embed resources/*
var embeddedFS embed.FS
