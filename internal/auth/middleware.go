package auth

import (
	"net/http"
	"strings"

	"github.com/aipass/aipass/internal/domain"
	"github.com/labstack/echo/v4"
)

const claimsKey = "auth_claims"

func Middleware(tokens *TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if tokens == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "auth_not_configured")
			}
			header := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing_bearer_token")
			}
			claims, err := tokens.Parse(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid_token")
			}
			c.Set(claimsKey, claims)
			return next(c)
		}
	}
}

func RequireRole(roles ...domain.Role) echo.MiddlewareFunc {
	allowed := map[domain.Role]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := ClaimsFromContext(c)
			if !ok || !allowed[claims.Role] {
				return echo.NewHTTPError(http.StatusForbidden, "forbidden")
			}
			return next(c)
		}
	}
}

func ClaimsFromContext(c echo.Context) (*Claims, bool) {
	claims, ok := c.Get(claimsKey).(*Claims)
	return claims, ok
}
