package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	inimodel "github.com/Proyek-Three/be-promosi-umkm/model"
	cek "github.com/Proyek-Three/be-promosi-umkm/module"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/aiteung/musik"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
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

// GetProductsByUserID handler untuk mendapatkan produk berdasarkan user_id
func GetProductsByUserID(c *fiber.Ctx) error {
	// Ambil user_id dari parameter URL
	userIDHex := c.Params("user_id")
	if userIDHex == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id is required",
		})
	}

	// Konversi user_id ke ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	// Panggil fungsi dari backend untuk mendapatkan produk
	products := cek.GetProductsByUserID(config.Ulbimongoconn, "product", userID)

	// Kembalikan JSON response
	return c.JSON(products)
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

	// Parse form data termasuk file gambar
	if err := c.BodyParser(&productdata); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	// Ambil file gambar dari request
	file, err := c.FormFile("image") // Asumsikan input file bernama "image"
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Failed to get image file: " + err.Error(),
		})
	}

	// Baca isi file
	imageFile, err := file.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to open image file: " + err.Error(),
		})
	}
	defer imageFile.Close()

	// Baca data file
	imageData, err := ioutil.ReadAll(imageFile)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to read image file: " + err.Error(),
		})
	}

	// Step 1: Upload gambar ke GitHub
	githubToken := os.Getenv("GH_ACCESS_TOKEN") // Ganti dengan token Anda
	repoOwner := "Proyek-Three"                 // Nama organisasi GitHub
	repoName := "images"                        // Nama repositori
	filePath := fmt.Sprintf("product/%d_%s.jpg", time.Now().Unix(), productdata.ProductName)
	uploadURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", repoOwner, repoName, filePath)

	encodedImage := base64.StdEncoding.EncodeToString(imageData)
	payload := map[string]string{
		"message": fmt.Sprintf("Add image for product %s", productdata.ProductName),
		"content": encodedImage,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", uploadURL, bytes.NewReader(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to upload image to GitHub: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "GitHub API error: " + string(body),
		})
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	imageURL := result["content"].(map[string]interface{})["download_url"].(string)

	// Tambahkan URL gambar ke data produk
	productdata.Image = imageURL

	// Step 2: Simpan data produk ke database
	insertedID, err := cek.InsertProduct(db, "product", productdata)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":      http.StatusOK,
		"message":     "Product data saved successfully.",
		"inserted_id": insertedID,
		"image_url":   imageURL,
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
