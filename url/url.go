package url

import (
	"github.com/Proyek-Three/bp-promosi-umkm/controller"
	"github.com/gofiber/fiber/v2"
)

func Web(page *fiber.App) {
	// page.Post("/api/whatsauth/request", controller.PostWhatsAuthRequest)  //API from user whatsapp message from iteung gowa
	// page.Get("/ws/whatsauth/qr", websocket.New(controller.WsWhatsAuthQR)) //websocket whatsauth

	page.Get("/", controller.Sink)
	page.Post("/", controller.Sink)
	page.Put("/", controller.Sink)
	page.Patch("/", controller.Sink)
	page.Delete("/", controller.Sink)
	page.Options("/", controller.Sink)
	page.Get("/checkip", controller.Homepage) //ujicoba panggil package musik
	// AUTH
	page.Post("/users/register", controller.Register)
	page.Post("/users/login", controller.Login)

	// PRODUCT
	page.Post("/insert/product", controller.InsertDataProduct)        //menambahkan data product
	page.Get("/product", controller.GetAllProduct)                    //menampilkan semua data product
	page.Get("/product/:id", controller.GetProductID)                 //menampilkan data product berdasarkan id
	page.Put("/update/product/:id", controller.UpdateDataProduct)     //update data product
	page.Delete("/product//delete/:id", controller.DeleteProductByID) //delete data product

	// page.Get("/docs/*", swagger.HandlerDefault)
}
