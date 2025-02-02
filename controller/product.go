package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
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

	// Parse data dari body
	if err := c.BodyParser(&productdata); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid input: " + err.Error(),
		})
	}

	// Validasi ID User
	if productdata.User.ID.IsZero() {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "User ID is required",
		})
	}
	var user inimodel.Users
	if err := db.Collection("users").FindOne(c.Context(), bson.M{"_id": productdata.User.ID}).Decode(&user); err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  http.StatusNotFound,
			"message": "User ID not found",
		})
	}
	if user.Store.StoreName == "" || user.Store.Address == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Store data is incomplete for the user.",
		})
	}
	productdata.User.Username = user.Username
	productdata.User.Store.StoreName = user.Store.StoreName
	productdata.User.Store.Address = user.Store.Address

	// Validasi ID Category
	if productdata.Category.ID.IsZero() {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Category ID is required",
		})
	}
	var category inimodel.Category
	if err := db.Collection("categories").FindOne(c.Context(), bson.M{"_id": productdata.Category.ID}).Decode(&category); err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  http.StatusNotFound,
			"message": "Category ID not found",
		})
	}
	productdata.Category.CategoryName = category.CategoryName

	// Validasi ID Status
	if productdata.Status.ID.IsZero() {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Status ID is required",
		})
	}
	var status inimodel.Status
	if err := db.Collection("statuses").FindOne(c.Context(), bson.M{"_id": productdata.Status.ID}).Decode(&status); err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  http.StatusNotFound,
			"message": "Status ID not found",
		})
	}
	productdata.Status.Status = status.Status

	// Proses upload gambar
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Image file is required: " + err.Error(),
		})
	}
	imageURL, err := UploadImageToGitHub(file, productdata.ProductName)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}
	productdata.Image = imageURL

	// Simpan data produk
	insertedID, err := InsertProduct(db, "product", productdata)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to save product: " + err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":      http.StatusOK,
		"message":     "Product data saved successfully.",
		"inserted_id": insertedID,
		"image_url":   imageURL,
	})
}

func UploadImageToGitHub(file *multipart.FileHeader, productName string) (string, error) {
	githubToken := os.Getenv("GH_ACCESS_TOKEN")
	repoOwner := "Proyek-Three"
	repoName := "images"
	filePath := fmt.Sprintf("product/%d_%s.jpg", time.Now().Unix(), productName)

	fileContent, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer fileContent.Close()

	imageData, err := ioutil.ReadAll(fileContent)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageData)
	payload := map[string]string{
		"message": fmt.Sprintf("Add image for product %s", productName),
		"content": encodedImage,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", repoOwner, repoName, filePath), bytes.NewReader(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error: %s", body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	content, ok := result["content"].(map[string]interface{})
	if !ok || content["download_url"] == nil {
		return "", fmt.Errorf("GitHub API response missing download_url")
	}

	return content["download_url"].(string), nil
}

func UpdateDataProduct(c *fiber.Ctx) error {
	db := config.Ulbimongoconn

	// Ambil ID produk dari parameter URL
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "ID produk tidak valid.",
		})
	}

	// Cek apakah produk sudah ada di database
	var existingProduct inimodel.Product
	err = db.Collection("product").FindOne(c.Context(), bson.M{"_id": objectID}).Decode(&existingProduct)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  http.StatusNotFound,
			"message": "Produk tidak ditemukan.",
		})
	}

	// Parse body request ke struct Product
	var product inimodel.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Gagal memproses request body.",
		})
	}

	// Validasi dan perbarui data Kategori
	if !product.Category.ID.IsZero() {
		var category inimodel.Category
		err := db.Collection("categories").FindOne(c.Context(), bson.M{"_id": product.Category.ID}).Decode(&category)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Kategori ID tidak ditemukan.",
			})
		}
		product.Category.CategoryName = category.CategoryName
	} else {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Kategori ID diperlukan.",
		})
	}

	// Tidak mengubah status jika status.ID tidak disertakan
	if !product.Status.ID.IsZero() {
		var status inimodel.Status
		err := db.Collection("statuses").FindOne(c.Context(), bson.M{"_id": product.Status.ID}).Decode(&status)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  http.StatusNotFound,
				"message": "Status ID tidak ditemukan.",
			})
		}
		product.Status.Status = status.Status
	} else {
		// Jika status.ID tidak ada, gunakan status yang sudah ada di database
		product.Status = existingProduct.Status
	}

	// Tidak mengubah ID Pengguna jika tidak disertakan
	// Tidak mengubah ID Pengguna jika tidak disertakan
	if product.User.ID.IsZero() {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "User ID is required",
		})
	}

	var user inimodel.Users
	if err := db.Collection("users").FindOne(c.Context(), bson.M{"_id": product.User.ID}).Decode(&user); err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  http.StatusNotFound,
			"message": "User ID not found",
		})
	}

	// Validasi store data
	if user.Store.StoreName == "" || user.Store.Address == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Store data is incomplete for the user.",
		})
	}

	// Update data pada objek product
	product.User.Username = user.Username
	product.User.Store.StoreName = user.Store.StoreName
	product.User.Store.Address = user.Store.Address

	fmt.Println("Data produk yang akan diperbarui: ", product)

	// Proses upload gambar
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Image file is required: " + err.Error(),
		})
	}
	imageURL, err := UploadImageToGitHub(file, product.ProductName)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}
	product.Image = imageURL

	// Update data produk di database
	err = UpdateProduct(db, "product", objectID, product)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}

	// Response sukses
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "Data produk berhasil diperbarui.",
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

// INSERT PRODUCT
func InsertProduct(db *mongo.Database, col string, product inimodel.Product) (insertedID primitive.ObjectID, err error) {
	// Validasi ID kategori
	if product.Category.ID.IsZero() {
		return primitive.NilObjectID, fmt.Errorf("invalid category ID: cannot be empty")
	}

	// Validasi ID status
	if product.Status.ID.IsZero() {
		return primitive.NilObjectID, fmt.Errorf("invalid status ID: cannot be empty")
	}

	// Menyusun dokumen BSON untuk produk
	productData := bson.M{
		"product_name": product.ProductName,
		"description":  product.Description,
		"image":        product.Image,
		"price":        product.Price,
		"category": bson.M{
			"_id":           product.Category.ID,
			"category_name": product.Category.CategoryName,
		},
		"status": bson.M{
			"_id":    product.Status.ID,
			"status": product.Status.Status,
		},
		"user": bson.M{
			"_id":      product.User.ID,
			"username": product.User.Username,
			"store": bson.M{
				"store_name": product.User.Store.StoreName,
				"address":    product.User.Store.Address,
			},
		},
	}

	// Menyisipkan dokumen ke MongoDB
	collection := db.Collection(col)
	result, err := collection.InsertOne(context.TODO(), productData)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to insert product: %w", err)
	}

	// Mendapatkan ID yang disisipkan
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("failed to parse inserted ID")
	}

	return insertedID, nil
}

func UpdateProduct(db *mongo.Database, col string, productID primitive.ObjectID, updatedProduct inimodel.Product) error {
	// Logging untuk debugging
	fmt.Printf("Updating product ID: %s with data: %+v\n", productID.Hex(), updatedProduct)

	// Validasi ID kategori
	if updatedProduct.Category.ID.IsZero() {
		return fmt.Errorf("invalid category ID: cannot be empty")
	}

	// Validasi ID status
	if updatedProduct.Status.ID.IsZero() {
		return fmt.Errorf("invalid status ID: cannot be empty")
	}

	// Menyusun dokumen BSON untuk pembaruan
	updateData := bson.M{
		"$set": bson.M{
			"product_name": updatedProduct.ProductName,
			"description":  updatedProduct.Description,
			"image":        updatedProduct.Image,
			"price":        updatedProduct.Price,
			"category": bson.M{
				"_id":           updatedProduct.Category.ID,
				"category_name": updatedProduct.Category.CategoryName,
			},
			"status": bson.M{
				"_id":    updatedProduct.Status.ID,
				"status": updatedProduct.Status.Status,
			},
			"user": bson.M{
				"_id":      updatedProduct.User.ID,
				"username": updatedProduct.User.Username,
				"store": bson.M{
					"store_name": updatedProduct.User.Store.StoreName,
					"address":    updatedProduct.User.Store.Address,
				},
			},
		},
	}

	// Mendapatkan koleksi dan memperbarui dokumen
	collection := db.Collection(col)
	filter := bson.M{"_id": productID}
	result, err := collection.UpdateOne(context.TODO(), filter, updateData)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	// Memastikan ada dokumen yang diperbarui
	if result.MatchedCount == 0 {
		return fmt.Errorf("no product found with ID: %s", productID.Hex())
	}

	fmt.Printf("Successfully updated product ID: %s\n", productID.Hex())
	return nil
}

func UpdateProductStatus(c *fiber.Ctx) error {
	db := config.Ulbimongoconn

	// Ambil ID produk dari parameter URL
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "ID produk tidak valid.",
		})
	}

	// Parse body request untuk mendapatkan status_id
	var request struct {
		StatusID string `json:"status_id"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Gagal memproses request body.",
		})
	}

	// Validasi StatusID
	statusObjectID, err := primitive.ObjectIDFromHex(request.StatusID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Status ID tidak valid.",
		})
	}

	// Validasi dan ambil data status dari database
	var status inimodel.Status
	err = db.Collection("statuses").FindOne(c.Context(), bson.M{"_id": statusObjectID}).Decode(&status)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  http.StatusNotFound,
			"message": "Status ID tidak ditemukan di database.",
		})
	}

	// Cek apakah produk dengan ID tersebut ada
	var product inimodel.Product
	err = db.Collection("product").FindOne(c.Context(), bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  http.StatusNotFound,
			"message": "Produk dengan ID tersebut tidak ditemukan.",
		})
	}

	// Update status produk di database
	update := bson.M{
		"$set": bson.M{
			"status.id":     status.ID,
			"status.status": status.Status,
		},
	}
	_, err = db.Collection("product").UpdateOne(c.Context(), bson.M{"_id": objectID}, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Gagal memperbarui status produk.",
		})
	}

	// Response sukses
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "Status produk berhasil diperbarui.",
	})
}
