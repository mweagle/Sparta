// +build !lambdabinary

package interceptor

import (
	"context"
	"encoding/json"
)

func (xri *xrayInterceptor) Begin(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}

func (xri *xrayInterceptor) BeforeSetup(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}
func (xri *xrayInterceptor) AfterSetup(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}
func (xri *xrayInterceptor) BeforeDispatch(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}
func (xri *xrayInterceptor) AfterDispatch(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}
func (xri *xrayInterceptor) Complete(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}
