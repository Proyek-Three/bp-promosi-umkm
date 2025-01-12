package config

import (
	"os"
	"time"
	"github.com/golang-jwt/jwt/v4"
	"github.com/Proyek-Three/be-promosi-umkm/model"
	"golang.org/x/crypto/bcrypt"
	
)

var MySigningKey = []byte(os.Getenv("lsjdflsdjfdsfsdioy45hahay"))


// GenerateJWT generates a JWT token for the admin
func GenerateJWT(admin model.Admin) (string, error) {
	claims := jwt.MapClaims{
		"admin_id": admin.ID.Hex(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(MySigningKey)
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
