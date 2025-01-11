package controller


import (
	"errors"
	"fmt"
	"net/http"
	inimodel "github.com/Proyek-Three/be-promosi-umkm/model"
	cek "github.com/Proyek-Three/be-promosi-umkm/module"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAllUsers(c *fiber.Ctx) error {
    users, err := cek.GetAllUsers(config.Ulbimongoconn, "users") 
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
            "status":  http.StatusInternalServerError,
            "message": err.Error(),
        })
    }
    return c.JSON(users)
}


func GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Wrong parameter",
		})
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid id parameter",
		})
	}

	user, err := cek.GetUserByID(objID, config.Ulbimongoconn, "users")
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": fmt.Sprintf("No data found for id %s", id),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": fmt.Sprintf("Error retrieving data for id %s", id),
		})
	}
	return c.JSON(user)
}

func InsertDataUser(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	var userData inimodel.DataUsers

	// Parse JSON input ke struct
	if err := c.BodyParser(&userData); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to parse request body.",
			"error":   err.Error(),
		})
	}

	// Ambil Store ID dari request body, jika ada
	storeID := userData.Store.ID

	// Panggil fungsi InsertUser (storeID bisa kosong)
	insertedID, err := cek.InsertUser(db, "users", userData, storeID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to insert user.",
			"error":   err.Error(),
		})
	}

	// Return response berhasil
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":      http.StatusOK,
		"message":     "User data saved successfully.",
		"inserted_id": insertedID.Hex(),
	})
}


func UpdateDataUser(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	var updatedUser inimodel.DataUsers

	// Parse JSON input ke struct
	if err := c.BodyParser(&updatedUser); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to parse request body.",
			"error":   err.Error(),
		})
	}

	// Cek apakah ID user diberikan
	userIDParam := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(userIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid user ID.",
			"error":   err.Error(),
		})
	}

	// Ambil Store ID dari request body, jika ada
	storeID := updatedUser.Store.ID

	// Panggil fungsi UpdateUser (storeID bisa kosong)
	err = cek.UpdateUser(db, "users", userID, updatedUser, storeID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to update user.",
			"error":   err.Error(),
		})
	}

	// Return response berhasil
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "User data updated successfully.",
		"user_id": userID.Hex(),
	})
}


func DeleteUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Wrong parameter",
		})
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid id parameter",
		})
	}

	err = cek.DeleteUserByID(objID, config.Ulbimongoconn, "users")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": fmt.Sprintf("Error deleting data for id %s", id),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": fmt.Sprintf("User data with id %s deleted successfully", id),
	})
}
