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
		dataRegis.Store.Address == "" || dataRegis.Store.Sosmed == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  http.StatusBadRequest,
			"message": "All fields are required",
		})
	}

	// Check if username or email already exists
	existingUser, err := cek.GetUserByUsernameOrEmail(config.Ulbimongoconn, "Users", dataRegis.Username, dataRegis.Email)
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
	insertedID, err := cek.RegisUser(config.Ulbimongoconn, "Users", dataRegis)
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
	existingUser, err := cek.GetUserByUsernameOrEmail(config.Ulbimongoconn, "Users", loginData.Username, loginData.Email)
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
