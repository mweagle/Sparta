package sparta

import (
	"path/filepath"
	"testing"

	spartaSystem "github.com/mweagle/Sparta/system"
	gocf "github.com/mweagle/go-cloudformation"
)

func TestS3Site(t *testing.T) {
	// Get the gopath, find this source directory and
	// make a site out of the documentation
	gopath := spartaSystem.GoPath()
	docsPath := filepath.Join(gopath, "src", "github.com", "mweagle", "Sparta", "docs")
	s3Site, s3SiteErr := NewS3Site(docsPath)
	if s3SiteErr != nil {
		t.Fatalf("Failed to create S3 Site: %s", s3SiteErr)
	}
	s3Site.BucketName = gocf.String("sparta-site.spartademo.net")
	apiStage := NewStage("v1")
	apiGateway := NewAPIGateway("SpartaTestSite", apiStage)

	testProvisionEx(t,
		testLambdaData(),
		apiGateway,
		s3Site,
		nil,
		false,
		nil,
	)
}
