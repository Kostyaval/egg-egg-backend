package rest

import "github.com/gofiber/fiber/v2"

func (h handler) ping(c *fiber.Ctx) error {
	log := h.log.HTTPRequest(c)
	log.Info("ping")

	return c.Status(fiber.StatusOK).Send(nil)
}
