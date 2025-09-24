package middleware

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/tg123/go-htpasswd"
)

type authKey struct{}

type AuthInfo struct {
	Username string
}

func WithUserInfo(ctx context.Context, authInfo *AuthInfo) context.Context {
	return context.WithValue(ctx, &authKey{}, authInfo)
}

func GetUserInfo(ctx context.Context) *AuthInfo {
	if v := ctx.Value(&authKey{}); v != nil {
		if authInfo, ok := v.(*AuthInfo); ok {
			return authInfo
		}
	}
	return nil
}

func Authorizer(authFile *htpasswd.File) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authorization := c.Request().Header.Get("Authorization")
			if authorization == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "No Authorization header"})
			}

			decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authorization, "Basic "))
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid authorization header format"})
			}

			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) != 2 {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid authorization header format"})
			}

			if !authFile.Match(parts[0], parts[1]) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid username or password"})
			}

			log.Ctx(c.Request().Context()).Debug().Str("user", parts[0]).Msg("Authenticated user")

			ctx := WithUserInfo(c.Request().Context(), &AuthInfo{
				Username: parts[0],
			})
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
