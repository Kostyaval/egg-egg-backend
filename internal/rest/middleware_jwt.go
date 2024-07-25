package rest

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
	"strings"
)

type middlewareJWTConfig struct {
	log          *handlerLogger
	cfg          *config.JWTConfig
	mustNickname bool
}

func middlewareJWT(mw *middlewareJWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := mw.log.HTTPRequest(c)

		// the middleware used twice, so no need to decode the jwt again
		_claims, ok := c.Locals("jwt").(*domain.JWTClaims)
		if ok {
			if mw.mustNickname && _claims.Nickname == nil {
				log.Error("no nickname", slog.Int64("uid", _claims.UID))
				return newHTTPError(fiber.StatusForbidden, "no nickname")
			}

			return c.Next()
		}

		token := strings.TrimPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
		if token == "" {
			return newHTTPError(fiber.StatusUnauthorized, "empty authorization header")
		}

		var claims domain.JWTClaims
		if err := claims.Decode(mw.cfg, []byte(token)); err != nil {
			log.Error(err.Error())
			return newHTTPError(fiber.StatusUnauthorized, "not authorized")
		}

		if mw.mustNickname && claims.Nickname == nil {
			log.Error("no nickname", slog.Int64("uid", claims.UID))
			return newHTTPError(fiber.StatusForbidden, "no nickname")
		}

		c.Locals("jwt", &claims)

		return c.Next()
	}
}
