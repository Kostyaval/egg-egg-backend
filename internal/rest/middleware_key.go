package rest

import (
	"github.com/gofiber/fiber/v2"
	"strings"
)

type middlewareAPIKeyConfig struct {
	log *handlerLogger
	key string
}

func middlewareAPIKey(mw *middlewareAPIKeyConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := mw.log.HTTPRequest(c)

		if mw.key == "" {
			log.Error("no token provided in config")
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		header := strings.TrimSpace(c.Get("X-Api-Key"))
		if header == "" {
			log.Warn("no X-Api-Key header provided")
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		if header != mw.key {
			log.Warn("invalid X-Api-Key header")
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		return c.Next()
	}
}
