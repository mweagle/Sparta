//go:build !lambdabinary
// +build !lambdabinary

package interceptor

import (
	"container/ring"
	"context"
	"encoding/json"

	sparta "github.com/mweagle/Sparta"
	"github.com/rs/zerolog"
)

func (xri *xrayInterceptor) Begin(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}

func (xri *xrayInterceptor) BeforeSetup(ctx context.Context, msg json.RawMessage) context.Context {
	return ctx
}
func (xri *xrayInterceptor) AfterSetup(ctx context.Context, msg json.RawMessage) context.Context {
	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*zerolog.Logger)
	if loggerOk {
		// Empty assignment to satisfy linters...
		xri.zerologXRayHandler = &zerologXRayHandler{
			logRing:       ring.New(logRingSize),
			originalLevel: logger.GetLevel(),
		}
	}
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
