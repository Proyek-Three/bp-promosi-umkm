package main

import (
	"log"

	"github.com/aiteung/musik"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	// _ "github.com/ghaidafasya24/bp-tubes/docs"
	"github.com/Proyek-Three/bp-promosi-umkm/url"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

)

func main() {
	site := fiber.New(config.Iteung)
	site.Use(cors.New(config.Cors))
	url.Web(site)
	log.Fatal(site.Listen(musik.Dangdut()))
	


}
