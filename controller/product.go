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
	db := config.Ulbimongoconn
	var productdata inimodel.Product

	// Parsing JSON input ke struct
	if err := c.BodyParser(&productdata); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
	}

	// Parsing ID dari parameter URL
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid ID format",
		})
	}

	// Validasi dan pembaruan kategori
	if !productdata.Category.ID.IsZero() {
		var category inimodel.Category
		err := db.Collection("categories").FindOne(c.Context(), bson.M{"_id": productdata.Category.ID}).Decode(&category)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Category ID not found.",
			})
		}
		productdata.Category.CategoryName = category.CategoryName
	} else {
		productdata.Category.ID = primitive.NewObjectID()
		productdata.Category.CategoryName = "Default Category"
	}

	// Validasi dan pembaruan toko
	if productdata.Store.ID.IsZero() {
		productdata.Store.ID = primitive.NewObjectID()
	}

	// Persiapan data untuk pembaruan
	update := bson.M{
		"$set": bson.M{
			"product_name":  productdata.ProductName,
			"description":   productdata.Description,
			"image":         productdata.Image,
			"price":         productdata.Price,
			"category": bson.M{
				"id":            productdata.Category.ID,
				"category_name": productdata.Category.CategoryName,
			},
			"store": bson.M{
				"id":      productdata.Store.ID,
				"store_name": productdata.Store.StoreName,
				"address": productdata.Store.Address,
			},
		},
	}

	// Melakukan pembaruan pada dokumen berdasarkan ID
	_, err = db.Collection("product").UpdateOne(c.Context(), bson.M{"_id": objectID}, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	// Mengembalikan respons berhasil
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "Product updated successfully.",
	})
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
