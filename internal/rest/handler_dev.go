package rest

import "github.com/gofiber/fiber/v2"

func (h handler) rules(c *fiber.Ctx) error {
	return c.JSON(h.cfg.Rules)
}
