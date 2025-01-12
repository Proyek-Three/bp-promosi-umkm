package middleware

import (
	"strings"
	"time"

	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware checks the validity of the JWT token
func JWTMiddleware(c *fiber.Ctx) error {
	// Extract token from the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is required",
		})
	}

	// Token should be in the format "Bearer <token>"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid signing method")
		}
		return config.MySigningKey, nil

	})

	// If the token is invalid
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// Extract claims (e.g., admin_id) from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	// Check expiration time
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token has expired",
			})
		}
	}

	// Store the admin_id in the context for further use
	c.Locals("admin_id", claims["admin_id"])

	// Continue to the next handler
	return c.Next()
}
