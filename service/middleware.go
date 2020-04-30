package service

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/go-kit/kit/endpoint"
)

func LoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				r := recover()
				logger.Log(
					"panic", r,
					"error", err,
					"req", request,
					"took", time.Since(begin))
			}(time.Now())
			return next(ctx, request)
		}
	}
}
