package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	inimodel "github.com/Proyek-Three/be-promosi-umkm/model"
	cek "github.com/Proyek-Three/be-promosi-umkm/module"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var jwtKey = []byte("secret_key!234@!#$%")

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func Register(c *fiber.Ctx) error {
	var dataRegis inimodel.Users

	// Parse body request
	if err := c.BodyParser(&dataRegis); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
	}

	// Validasi semua field harus terisi
	if dataRegis.Name == "" || dataRegis.Username == "" || dataRegis.Password == "" ||
		dataRegis.Email == "" || dataRegis.PhoneNumber == "" || dataRegis.Store.StoreName == "" ||
		dataRegis.Store.Address == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "All fields are required",
		})
	}

	// Check if username or email already exists
	existingUser, err := cek.GetUserByUsernameOrEmail(config.Ulbimongoconn, "users", dataRegis.Username, dataRegis.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Could not check existing username or email",
		})
	}
	if existingUser != nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"status":  http.StatusConflict,
			"message": "Username or email already taken",
		})
	}

	// Set default role to "seller"
	dataRegis.Role = "seller"

	// Save user with store to database
	insertedID, err := cek.RegisUser(config.Ulbimongoconn, "users", dataRegis)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Could not register user",
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"status":  http.StatusCreated,
		"message": "Account registered successfully",
		"data":    insertedID,
	})
}

func GetUserProfile(c *fiber.Ctx) error {
	// Ambil username dari query parameter atau header (sesuaikan dengan kebutuhan Anda)
	username := c.Query("username")
	if username == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Username is required",
		})
	}

	// Ambil data pengguna dari database berdasarkan username
	user, err := cek.GetUserByUsername(config.Ulbimongoconn, "users", username)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to retrieve user profile",
		})
	}
	if user == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  http.StatusNotFound,
			"message": "User not found",
		})
	}

	// Kirim data pengguna ke frontend (tanpa menyertakan password untuk keamanan)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "User profile retrieved successfully",
		"data": fiber.Map{
			"name":         user.Name,
			"username":     user.Username,
			"email":        user.Email,
			"phone_number": user.PhoneNumber,
			"store": fiber.Map{
				"store_name": user.Store.StoreName,
				"address":    user.Store.Address,
			},
		},
	})
}

func Login(c *fiber.Ctx) error {
	// Parse login data from request body
	var loginData inimodel.Users
	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
	}

	// Validasi bahwa username/email dan password tidak kosong
	if loginData.Username == "" && loginData.Email == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Username or email is required",
		})
	}
	if loginData.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Password is required",
		})
	}

	// Ambil user berdasarkan username atau email
	existingUser, err := cek.GetUserByUsernameOrEmail(config.Ulbimongoconn, "users", loginData.Username, loginData.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Error fetching user data",
		})
	}
	if existingUser == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Invalid username, email, or password",
		})
	}

	// Validasi password
	if !cek.ValidatePassword(existingUser.Password, loginData.Password) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Invalid username,or password",
		})
	}

	// Buat JWT token
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		UserID:   existingUser.ID,
		Username: existingUser.Username,
		Email:    existingUser.Email,
		Role:     existingUser.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Error generating token",
		})
	}

	// Berhasil login
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "Login successful",
		"token":   tokenString, // Token dikembalikan ke client
		"data": fiber.Map{
			"role": existingUser.Role,
		},
	})
}

// ValidateToken memvalidasi token JWT
func ValidateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return false, err
	}
	return token.Valid, nil
}

// JWTAuth middleware untuk memverifikasi token di Fiber
func JWTAuth(c *fiber.Ctx) error {
	bearerToken := c.Get("Authorization") // Ambil Authorization header
	sttArr := strings.Split(bearerToken, " ")
	if len(sttArr) == 2 {
		isValid, _ := ValidateToken(sttArr[1]) // Validasi token
		if isValid {
			return c.Next() // Lanjutkan ke handler berikutnya jika token valid
		}
	}
	return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
		"message": "Unauthorized",
	}) // Jika tidak valid
}

func GetAllUser(c *fiber.Ctx) error {
	collection := config.Ulbimongoconn.Collection("users")
	users, err := cek.GetAllUser(collection)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Error fetching users",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": http.StatusOK,
		"data":   users,
	})
}

func GetProfile(c *fiber.Ctx) error {
	bearerToken := c.Get("Authorization")
	sttArr := strings.Split(bearerToken, " ")
	if len(sttArr) != 2 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Unauthorized",
		})
	}

	tokenString := sttArr[1]
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Invalid token",
		})
	}

	userID := claims.UserID
	collection := config.Ulbimongoconn.Collection("users")
	user, err := cek.GetUsersByID(collection, userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Error fetching profile",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": http.StatusOK,
		"data":   user,
	})
}

func UpdateUser(c *fiber.Ctx) error {
	bearerToken := c.Get("Authorization")
	sttArr := strings.Split(bearerToken, " ")
	if len(sttArr) != 2 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Unauthorized",
		})
	}

	tokenString := sttArr[1]
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Invalid token",
		})
	}

	// Konversi UserID dari string ke ObjectID
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Invalid UserID format",
		})
	}

	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
	}

	// Debugging: Print request body
	fmt.Println("Request Body:", updateData)

	// Update user
	updatedUser, err := cek.UpdateUser(userCollection, userID, updateData)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Failed to update profile",
			"error":   err.Error(),
		})
	}

	// Auto-Update store dan username di Collection Products
	updateFields := bson.M{}

	// Update store details
	if storeData, ok := updateData["store"].(map[string]interface{}); ok {
		if storeName, hasStoreName := storeData["store_name"].(string); hasStoreName {
			updateFields["user.store.store_name"] = storeName
		}
		if storeAddress, hasStoreAddress := storeData["address"].(string); hasStoreAddress {
			updateFields["user.store.address"] = storeAddress
		}
	}

	// Update username jika ada perubahan
	if username, hasUsername := updateData["username"].(string); hasUsername {
		updateFields["user.username"] = username
	}

	// Debugging: Print updateFields
	fmt.Println("Update Fields:", updateFields)

	if len(updateFields) > 0 {
		productFilter := bson.M{"user._id": userID}
		productUpdate := bson.M{"$set": updateFields}

		// Debugging: Print query update
		fmt.Println("Updating Products with Filter:", productFilter)
		fmt.Println("Updating Products with Data:", productUpdate)

		_, err := productCollection.UpdateMany(context.Background(), productFilter, productUpdate)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  http.StatusInternalServerError,
				"message": "Failed to update products store details",
				"error":   err.Error(),
			})
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "Profile and related products updated successfully",
		"data":    updatedUser,
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

var userCollection *mongo.Collection
var productCollection *mongo.Collection

func InitCollections() {
	if config.Ulbimongoconn == nil {
		log.Fatal("Database connection is not initialized")
	}

	userCollection = config.Ulbimongoconn.Collection("users")
	productCollection = config.Ulbimongoconn.Collection("product")
}
