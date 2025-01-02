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

// GetAllStores retrieves all store data.
func GetAllStores(c *fiber.Ctx) error {
	stores := cek.GetAllStore(config.Ulbimongoconn, "stores")
	fmt.Println("Data yang akan dikirim: ", stores) // Debug log
	return c.JSON(stores)
}

// GetStoreByID retrieves store data by its ID.
func GetStoreByID(c *fiber.Ctx) error {
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

	store, err := cek.GetStoreFromID(objID, config.Ulbimongoconn, "stores")
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

	return c.JSON(store)
}

// InsertStore inserts a new store into the database.
func InsertStore(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	var store inimodel.Store

	if err := c.BodyParser(&store); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	insertedID, err := cek.InsertStore(db, "stores", store)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":      http.StatusOK,
		"message":     "Store data saved successfully.",
		"inserted_id": insertedID,
	})
}

// UpdateStore updates an existing store's data.
func UpdateStore(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	id := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	var payload struct {
		StoreName string `json:"store_name"`
		Address   string `json:"address"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	err = cek.UpdateStore(db, "stores", objectID, payload.StoreName, payload.Address)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "Store data successfully updated.",
	})
}

// DeleteStoreByID deletes a store from the database by its ID.
func DeleteStoreByID(c *fiber.Ctx) error {
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

	err = cek.DeleteStoreByID(objID, config.Ulbimongoconn, "stores")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": fmt.Sprintf("Error deleting data for id %s", id),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": fmt.Sprintf("Store data with id %s deleted successfully.", id),
	})
}
