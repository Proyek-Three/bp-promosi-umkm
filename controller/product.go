package controller

import (
	"fmt"
	"net/http"

	inimodel "github.com/Proyek-Three/be-promosi-umkm/model"
	cek "github.com/Proyek-Three/be-promosi-umkm/module"
	"github.com/Proyek-Three/bp-promosi-umkm/config"
	"github.com/aiteung/musik"
	"github.com/gofiber/fiber/v2"
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

func InsertDataProduct(c *fiber.Ctx) error {
	db := config.Ulbimongoconn
	var productdata inimodel.Product
	if err := c.BodyParser(&productdata); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}
	insertedID, err := cek.InsertProduct(db, "product", productdata)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":      http.StatusOK,
		"message":     "Data product berhasil disimpan.",
		"inserted_id": insertedID,
	})
}
