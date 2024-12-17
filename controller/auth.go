package controller

import (
	"net/http"

	inimodel "github.com/Proyek-Three/be-promosi-umkm/model"
	cek "github.com/Proyek-Three/be-promosi-umkm/module"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/gofiber/fiber/v2"
)

func Register(c *fiber.Ctx) error {
	var newAdmin inimodel.User
	if err := c.BodyParser(&newAdmin); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
	}

	// Check if username already exists
	existingAdmin, err := cek.GetAdminByUsername(config.Ulbimongoconn, "Admin", newAdmin.Username)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Could not check existing username",
		})
	}
	if existingAdmin != nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"status":  http.StatusConflict,
			"message": "Username already taken",
		})
	}

	// Save admin to database
	insertedID, err := cek.InsertAdmin(config.Ulbimongoconn, "Admin", newAdmin.Username, newAdmin.Password, newAdmin.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Could not register admin",
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"status":  http.StatusCreated,
		"message": "Account registered successfully",
		"data":    insertedID,
	})
}
