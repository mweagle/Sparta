//go:build lambdabinary

//lint:file-ignore U1000,ST1018 Ignore all unused code, it's generated
/* #nosec */

package sparta

import (
	"embed"
)

// content holds our static web server content.

//go:embed resources/awsbinary/README.md
var embeddedFS embed.FS
