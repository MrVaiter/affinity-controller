package routes

import (
    "github.com/gofiber/fiber/v2"

    "github.com/advantiss/pkg/controllers"
)

// PublicRoutes func for describe group of public routes.
func PublicRoutes(a *fiber.App) {
    // Create routes group.
    route := a.Group("/api/v1")

    // upload files
    route.Post("/deccontrol", controllers.ChangeConfig)
}
