package config

import (
	"github.com/gofiber/fiber/v2"
)

var Iteung = fiber.Config{
	Prefork:       false,
	CaseSensitive: true,
	StrictRouting: true,
	ServerHeader:  "Iteung",
	AppName:       "Message Router",

	
}

func InitServer() *fiber.App {
	return fiber.New()
}
func main() {
	app := fiber.New()

	// Jalankan server
	app.Listen(":5501")
}