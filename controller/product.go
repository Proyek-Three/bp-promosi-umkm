package controller

import (
	"errors"
	"fmt"
	"net/http"
	inimodel "github.com/Proyek-Three/be-promosi-umkm/model"
	cek "github.com/Proyek-Three/be-promosi-umkm/module"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/aiteung/musik"
	"go.mongodb.org/mongo-driver/bson"
	
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Homepage(c *fiber.Ctx) error {
	ipaddr := musik.GetIPaddress()
	return c.JSON(ipaddr)
}

func GetAllProduct(c *fiber.Ctx) error {
	ps := cek.GetAllProduct(config.Ulbimongoconn, "product")
	fmt.Println("Data yang akan dikirim: ", ps) // Tambahkan log ini
	return c.JSON(ps)
}

func GetProductID(c *fiber.Ctx) error {
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
	ps, err := cek.GetProductFromID(objID, config.Ulbimongoconn, "product")
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
	return c.JSON(ps)
}

func InsertDataProduct(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	var productdata inimodel.Product

	// Parse JSON input ke struct
	if err := c.BodyParser(&productdata); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	// Cek apakah kategori ID diberikan
	if !productdata.Category.ID.IsZero() {
		// Cari kategori berdasarkan ID
		var category inimodel.Category
		err := db.Collection("categories").FindOne(c.Context(), bson.M{"_id": productdata.Category.ID}).Decode(&category)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Category ID not found.",
			})
		}

		// Set category_name berdasarkan hasil pencarian
		productdata.Category.CategoryName = category.CategoryName
	} else {
		// Jika tidak ada ID kategori, buat ID baru dan tambahkan nama kategori default
		productdata.Category.ID = primitive.NewObjectID()
		productdata.Category.CategoryName = "Default Category"
	}

	// Generate ObjectID baru untuk toko jika ID tidak ada
	if productdata.Store.ID.IsZero() {
		productdata.Store.ID = primitive.NewObjectID()
	}

	// Insert data produk ke database
	insertedID, err := cek.InsertProduct(db, "product", productdata)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	// Return response berhasil
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":      http.StatusOK,
		"message":     "Product data saved successfully.",
		"inserted_id": insertedID.Hex(),
		"category_id": productdata.Category.ID.Hex(),
		"store_id":    productdata.Store.ID.Hex(),
		"category_name": productdata.Category.CategoryName,
	})
}



func UpdateProduct(c *fiber.Ctx) error {
	var input struct {
		ProductName  string            `json:"product_name" binding:"required"`
		Description  string            `json:"description" binding:"required"`
		Image        string            `json:"image" binding:"required"`
		Price        float64           `json:"price" binding:"required"`
		CategoryName inimodel.Category    `json:"category_name" binding:"required"`
		StoreName    inimodel.Store       `json:"store_name" binding:"required"`
		Address      inimodel.Store       `json:"address" binding:"required"`
	}

	// Binding the incoming JSON request to the struct
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Parsing ID from URL
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	// Get MongoDB database connection (you might already have a method for this)
	db := config.Ulbimongoconn // Replace with your actual method to get the DB connection

	// Call the repository function to update the product (passing the database instance)
	err = cek.UpdateProduct(db, "product", objectID, input.ProductName, input.Description, input.Image, input.Price, input.CategoryName, input.StoreName, input.Address)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Return success response
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Product updated successfully"})
}




func DeleteProductByID(c *fiber.Ctx) error {
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

	err = cek.DeleteProductByID(objID, config.Ulbimongoconn, "product")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": fmt.Sprintf("Error deleting data for id %s", id),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": fmt.Sprintf("Product data with id %s deleted successfully", id),
	})
}
