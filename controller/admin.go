package controller

import (
	"net/http"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"fmt"
	"strings"
	"github.com/Proyek-Three/be-promosi-umkm/model"
	"github.com/Proyek-Three/be-promosi-umkm/module"
	"github.com/gofiber/fiber/v2"
)

// LoginAdmin handles admin login requests.
func LoginAdmin(c *fiber.Ctx) error {
	var logindetails model.Admin

	// Parse the incoming JSON body into the Admin struct
	if err := c.BodyParser(&logindetails); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Fetch the admin from the database by username
	storeAdmin, err := module.GetAdminByUsername(config.Ulbimongoconn, logindetails.UserName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Invalid credentials",
		})
	}
	if storeAdmin == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Admin not found",
		})
	}

	// Check if the password is valid
	if !config.CheckPasswordHash(logindetails.Password, storeAdmin.Password) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Invalid credentials",
		})
	}

	// Generate the JWT token
	token, err := config.GenerateJWT(*storeAdmin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to generate token",
		})
	}

	// Save the token to the database
	err = module.SaveTokenToDatabase(config.Ulbimongoconn, "tokens", storeAdmin.ID.Hex(), token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to save token",
		})
	}

	// Return the generated token
	return c.JSON(fiber.Map{
		"message": "Login successful",
		"status":  "success",
		"token":   token,
	})
}

func RegisterAdmin(c *fiber.Ctx) error {
	var registerDetails model.Admin

	// Parse the incoming JSON body into the Admin struct
	if err := c.BodyParser(&registerDetails); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Check if the username is already taken
	existingAdmin, err := module.GetAdminByUsername(config.Ulbimongoconn, registerDetails.UserName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Error checking username",
		})
	}
	if existingAdmin != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Username already taken",
		})
	}

	// Hash the password before storing it
	hashedPassword, err := config.HashPassword(registerDetails.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to hash password",
		})
	}
	registerDetails.Password = hashedPassword

	// Save the new admin to the database
	insertedID, err := module.SaveAdminToDatabase(config.Ulbimongoconn, "adminumkm", registerDetails.UserName, registerDetails.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to save admin to database",
		})
	}

	// Return the success response
	return c.JSON(fiber.Map{
		"message": "Registration successful",
		"status":  "success",
		"id":      insertedID, // Return the inserted admin ID
	})
}

func Logout(c *fiber.Ctx) error {
	// Get the Authorization header from the request
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Missing token",
		})
	}

	// Split the Authorization header to extract the token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token format",
		})
	}

	token := parts[1]

	// Call the DeleteTokenFromMongoDB function to remove the token from the database
	err := module.DeleteTokenFromMongoDB(config.Ulbimongoconn, token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not delete token",
		})
	}

	// Return success message after token deletion
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logout successful",
	})
}

func DashboardPage(c *fiber.Ctx) error {
    adminID := c.Locals("admin_id")
    if adminID == nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
            "status":  http.StatusInternalServerError,
            "message": "Admin ID not found in context",
        })
    }

    adminIDStr := fmt.Sprintf("%v", adminID)

    return c.Status(http.StatusOK).JSON(fiber.Map{
        "status":  http.StatusOK,
        "message": "Dashboard access successful",
        "admin_id": adminIDStr,
    })
}