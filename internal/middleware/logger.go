package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)

			ctx := log.Logger.WithContext(c.Request().Context())

			l := log.Ctx(ctx)
			l.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("request_id", reqID)
			})

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
