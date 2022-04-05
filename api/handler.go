package api

import (
	"github.com/gofiber/fiber/v2"		
)

func middleware(c *fiber.Ctx) error {
	return c.Next()
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}

func info(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name": "Minion",
		"desc": "Thumbnailer for images and videos uploaded to minio",
		"version": "1.0.0",
		"tech-used": "go 1.17",
	})
}
