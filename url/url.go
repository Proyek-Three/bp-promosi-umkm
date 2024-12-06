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
	page.Post("/users/register", controller.Register)
	

	page.Get("/checkip", controller.Homepage) //ujicoba panggil package musik
	page.Get("/product", controller.GetAllProduct)
	// page.Get("/product/:id", controller.GetMenuID) //menampilkan data menu berdasarkan id
	page.Post("/insert/product", controller.InsertDataProduct)
	// page.Put("/update/:id", controller.UpdateData)
	// page.Delete("/delete/:id", controller.DeleteMenuByID)

	// page.Get("/docs/*", swagger.HandlerDefault)
}
