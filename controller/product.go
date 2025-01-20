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
	"strings"
	"time"

	inimodel "github.com/Proyek-Three/be-promosi-umkm/model"
	cek "github.com/Proyek-Three/be-promosi-umkm/module"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/aiteung/musik"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
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

// Fungsi untuk mengambil produk berdasarkan user_id
func GetProductsByUser(c *fiber.Ctx) error {
	// Ambil token dari header Authorization
	bearerToken := c.Get("Authorization")
	sttArr := strings.Split(bearerToken, " ")
	if len(sttArr) != 2 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Validasi token JWT
	token, err := jwt.Parse(sttArr[1], func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token",
		})
	}

	// Ambil user_id dari claims di dalam token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Ambil user_id sebagai string dan konversi ke ObjectID
	userIDStr := claims["user_id"].(string)
	userIDObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid user ID format",
		})
	}

	// Ambil data produk berdasarkan user._id (perhatikan user._id adalah ObjectID)
	db := config.Ulbimongoconn // Pastikan Anda sudah mengonfigurasi database
	var products []inimodel.Product
	cursor, err := db.Collection("product").Find(c.Context(), bson.M{
		"user._id": userIDObjectID, // Gunakan ObjectID dalam query MongoDB
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch products: " + err.Error(),
		})
	}
	defer cursor.Close(c.Context()) // Pastikan cursor ditutup

	// Mendekode setiap produk dalam cursor
	for cursor.Next(c.Context()) {
		var product inimodel.Product
		if err := cursor.Decode(&product); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to decode product: " + err.Error(),
			})
		}

		// Hanya ambil id dan username dari User
		product.User = inimodel.Users{
			ID:       product.User.ID,
			Username: product.User.Username,
		}

		// Menambahkan produk yang telah diubah ke dalam list
		products = append(products, product)
	}

	// Cek jika cursor error
	if err := cursor.Err(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Cursor error: " + err.Error(),
		})
	}

	// Jika produk tidak ditemukan
	if len(products) == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "No products found for this user",
		})
	}

	// Kembalikan data produk dalam bentuk JSON
	return c.JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "Products fetched successfully",
		"data":    products,
	})
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

	// Validasi field wajib
	if productdata.ProductName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Product name cannot be empty",
		})
	}

	if productdata.Description == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Product description cannot be empty",
		})
	}

	if productdata.Price <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Product price must be greater than zero",
		})
	}

	// Validasi kategori ID
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
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Category ID is required",
		})
	}

	// Validasi status ID
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
			"message": "Status ID is required",
		})
	}

	// Validasi user ID dan ambil Store ID dari koleksi Users
	if !productdata.User.ID.IsZero() {
		var user inimodel.Users
		err := db.Collection("users").FindOne(c.Context(), bson.M{"_id": productdata.User.ID}).Decode(&user)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "User ID not found.",
			})
		}
	
		if user.Store.ID.IsZero() {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"status":  http.StatusBadRequest,
				"message": "Store ID is missing for the user.",
			})
		}
	
		productdata.User.Username = user.Username
		productdata.StoreName = user.Store.StoreName
		productdata.StoreAddress = user.Store.Address
	} else {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "User ID is required.",
		})
	}
	

	// Proses upload gambar tetap sama
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Failed to get image file: " + err.Error(),
		})
	}

	imageFile, err := file.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to open image file: " + err.Error(),
		})
	}
	defer imageFile.Close()

	imageData, err := ioutil.ReadAll(imageFile)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to read image file: " + err.Error(),
		})
	}

	githubToken := os.Getenv("GH_ACCESS_TOKEN")
	repoOwner := "Proyek-Three"
	repoName := "images"
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
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to parse GitHub API response: " + err.Error(),
		})
	}
	content, ok := result["content"].(map[string]interface{})
	if !ok || content["download_url"] == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "GitHub API response missing download_url",
		})
	}
	imageURL := content["download_url"].(string)
	productdata.Image = imageURL

	// Simpan data produk ke database
	insertedID, err := cek.InsertProduct(db, "product", productdata)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to insert product: " + err.Error(),
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

	// Validasi User ID
	if !updatedProduct.User.ID.IsZero() {
		var user inimodel.Users
		err := db.Collection("users").FindOne(c.Context(), bson.M{"_id": updatedProduct.User.ID}).Decode(&user)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "User ID not found.",
			})
		}

		// Validasi apakah user memiliki store
		if user.Store.ID.IsZero() {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"status":  http.StatusBadRequest,
				"message": "User does not have a valid store ID.",
			})
		}

		// Ambil detail store dari koleksi stores
		var store inimodel.Store
		err = db.Collection("stores").FindOne(c.Context(), bson.M{"_id": user.Store.ID}).Decode(&store)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Store ID not found.",
			})
		}

		// Set store_name dan store_address dari koleksi stores
		updatedProduct.User.Username = user.Username
		updatedProduct.StoreName = store.StoreName
		updatedProduct.StoreAddress = store.Address
	} else {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "User ID is required.",
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
		"status_id":     updatedProduct.Status.ID.Hex(),
		"user_id":       updatedProduct.User.ID.Hex(),
		"category_name": updatedProduct.Category.CategoryName,
		"status_name":   updatedProduct.Status.Status,
		"store_name":    updatedProduct.StoreName,
		"store_address": updatedProduct.StoreAddress,
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
