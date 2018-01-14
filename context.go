package sparta

type contextKey int

const (
	// ContextKeyLogger is the request-independent *logrus.Logger
	// instance common to all requests
	ContextKeyLogger contextKey = iota
	// ContextKeyRequestLogger is the *logrus.Entry instance
	// that is annotated with request-identifying
	// information extracted from the AWS context object
	ContextKeyRequestLogger
	// ContextKeyLambdaContext is the *sparta.LambdaContext
	// pointer in the request
	// DEPRECATED
	ContextKeyLambdaContext
)

const (
	// ContextKeyLambdaVersions is the key in the context that stores the map
	// of autoincrementing versions
	ContextKeyLambdaVersions = "spartaLambdaVersions"
)
