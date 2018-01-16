// +build lambdabinary

package sparta

// NewS3Site returns a new S3Site pointer initialized with the
// static resources at the supplied path.  If resources is a directory,
// the contents will be recursively archived and used to populate
// the new S3 bucket.
func NewS3Site(resources string) (*S3Site, error) {
	return &S3Site{}, nil
}
