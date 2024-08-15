package rest

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"strings"
)

type middlewareJWTConfig struct {
	log *handlerLogger
	cfg *config.JWTConfig
}

func middlewareJWT(mw *middlewareJWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := mw.log.HTTPRequest(c)

		token := strings.TrimPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
		if token == "" {
			return newHTTPError(fiber.StatusUnauthorized, "empty authorization header")
		}

		claims, err := mw.cfg.Decode([]byte(token))
		if err != nil {
			log.Error(err.Error())
			return newHTTPError(fiber.StatusUnauthorized, "not authorized")
		}

		c.Locals("jwt", &claims)

		return c.Next()
	}
}
