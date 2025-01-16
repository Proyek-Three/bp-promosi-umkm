package controller

import (
	"net/http"
	"time"

	inimodel "github.com/Proyek-Three/be-promosi-umkm/model"
	cek "github.com/Proyek-Three/be-promosi-umkm/module"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
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
	var newAdmin inimodel.User
	if err := c.BodyParser(&newAdmin); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
	}

	// Check if username already exists
	existingAdmin, err := cek.GetUserByUsernameOrEmail(config.Ulbimongoconn, "Users", newAdmin.Username, newAdmin.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Could not check existing username",
		})
	}
	if existingAdmin != nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"status":  http.StatusConflict,
			"message": "Username already taken",
		})
	}

	// Save admin to database
	insertedID, err := cek.InsertAdmin(config.Ulbimongoconn, "Users", newAdmin.Username, newAdmin.Password, newAdmin.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Could not register admin",
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"status":  http.StatusCreated,
		"message": "Account registered successfully",
		"data":    insertedID,
	})
}

func Login(c *fiber.Ctx) error {
	var loginData inimodel.User
	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
	}

	// Fetch user by username or email
	existingAdmin, err := cek.GetUserByUsernameOrEmail(config.Ulbimongoconn, "Users", loginData.Username, loginData.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": "Error fetching user data",
		})
	}
	if existingAdmin == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Invalid credentials",
		})
	}

	// Validate password
	if !cek.ValidatePassword(existingAdmin.Password, loginData.Password) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  http.StatusUnauthorized,
			"message": "Invalid credentials",
		})
	}

	// Generate token
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		UserID:   existingAdmin.ID,
		Username: existingAdmin.Username,
		Email:    existingAdmin.Email,
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

	// Successful login
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  http.StatusOK,
		"message": "Login successful",
		"token":   tokenString, // Token ditambahkan di sini
	})
}
