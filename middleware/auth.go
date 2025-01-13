package middleware

import (
	"strings"
	"time"

	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gofiber/fiber/v2"
)


func JWTMiddleware(c *fiber.Ctx) error {
    // Jika path adalah Swagger UI atau Docs, tambahkan Authorization header jika belum ada
    if strings.HasPrefix(c.Path(), "/swagger.yaml") || strings.HasPrefix(c.Path(), "/docs/*") {
        if c.Get("Authorization") == "" {
            c.Request().Header.Set("Authorization", "Bearer lsjdflsdjfdsfsdioy45hahay")
        }
        return c.Next() // Lanjutkan ke handler berikutnya
    }

    // Extract token dari header Authorization
    authHeader := c.Get("Authorization")
    if authHeader == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Authorization header is required",
        })
    }

    // Token harus dalam format "Bearer <token>"
    tokenString := strings.TrimPrefix(authHeader, "Bearer ")

    // Parse token JWT
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Periksa metode signing
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid signing method")
        }
        return config.MySigningKey, nil
    })

    // Jika token tidak valid
    if err != nil || !token.Valid {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid or expired token",
        })
    }

    // Ekstrak klaim dari token (misalnya admin_id)
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid token claims",
        })
    }

    // Periksa waktu kedaluwarsa token
    if exp, ok := claims["exp"].(float64); ok {
        if time.Unix(int64(exp), 0).Before(time.Now()) {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Token has expired",
            })
        }
    }

    // Simpan admin_id di konteks untuk digunakan di handler berikutnya
    c.Locals("admin_id", claims["admin_id"])

    // Lanjutkan ke handler berikutnya
    return c.Next()
}


