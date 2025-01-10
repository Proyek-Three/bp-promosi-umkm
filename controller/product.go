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

	// Validasi Category ID
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

	// Validasi Store ID
	if !productdata.Store.ID.IsZero() {
		var store inimodel.Store
		err := db.Collection("stores").FindOne(c.Context(), bson.M{"_id": productdata.Store.ID}).Decode(&store)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Store ID not found.",
			})
		}
		productdata.Store.StoreName = store.StoreName
		productdata.Store.Address = store.Address
	} else {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Store ID is required.",
		})
	}

	// Validasi Status ID
	if !productdata.Status.ID.IsZero() {
		var status inimodel.Status
		err := db.Collection("statuses").FindOne(c.Context(), bson.M{"_id": productdata.Status.ID}).Decode(&status)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Status ID not found.",
			})
		}
		productdata.Status.Status = status.Status
	} else {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Status ID is required.",
		})
	}

	// Insert produk ke database
	insertedID, err := cek.InsertProduct(db, "product", productdata)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	// Return response berhasil
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":        http.StatusOK,
		"message":       "Product data saved successfully.",
		"inserted_id":   insertedID.Hex(),
		"category_id":   productdata.Category.ID.Hex(),
		"store_id":      productdata.Store.ID.Hex(),
		"status_id":     productdata.Status.ID.Hex(),
		"category_name": productdata.Category.CategoryName,
		"store_name":    productdata.Store.StoreName,
		"address":       productdata.Store.Address,
		"status_name":   productdata.Status.Status,
	})
}





func UpdateDataProduct(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	var updatedProduct inimodel.Product

	// Parse JSON input ke struct
	if err := c.BodyParser(&updatedProduct); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	// Validasi ID produk
	productIDParam := c.Params("id")
	productID, err := primitive.ObjectIDFromHex(productIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid product ID.",
		})
	}

	// Validasi Category ID
	if !updatedProduct.Category.ID.IsZero() {
		var category inimodel.Category
		err := db.Collection("categories").FindOne(c.Context(), bson.M{"_id": updatedProduct.Category.ID}).Decode(&category)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Category ID not found.",
			})
		}
		updatedProduct.Category.CategoryName = category.CategoryName
	} else {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Category ID is required.",
		})
	}

	// Validasi Store ID
	if !updatedProduct.Store.ID.IsZero() {
		var store inimodel.Store
		err := db.Collection("stores").FindOne(c.Context(), bson.M{"_id": updatedProduct.Store.ID}).Decode(&store)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Store ID not found.",
			})
		}
		updatedProduct.Store.StoreName = store.StoreName
		updatedProduct.Store.Address = store.Address
	} else {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Store ID is required.",
		})
	}

	// Validasi Status ID
	if !updatedProduct.Status.ID.IsZero() {
		var status inimodel.Status
		err := db.Collection("statuses").FindOne(c.Context(), bson.M{"_id": updatedProduct.Status.ID}).Decode(&status)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Status ID not found.",
			})
		}
		updatedProduct.Status.Status = status.Status
	} else {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Status ID is required.",
		})
	}

	// Update data produk
	err = cek.UpdateProduct(db, "product", productID, updatedProduct)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	// Return response berhasil
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":        http.StatusOK,
		"message":       "Product data updated successfully.",
		"product_id":    productID.Hex(),
		"category_id":   updatedProduct.Category.ID.Hex(),
		"store_id":      updatedProduct.Store.ID.Hex(),
		"status_id":     updatedProduct.Status.ID.Hex(),
		"category_name": updatedProduct.Category.CategoryName,
		"store_name":    updatedProduct.Store.StoreName,
		"address":       updatedProduct.Store.Address,
		"status_name":   updatedProduct.Status.Status,
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
