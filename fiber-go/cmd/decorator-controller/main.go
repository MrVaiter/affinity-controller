package main

import (
	"github.com/advantiss/pkg/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	routes.PublicRoutes(app)

	app.Listen(":3000")
}
