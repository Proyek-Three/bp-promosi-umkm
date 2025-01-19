package url

import (
	"github.com/Proyek-Three/bp-promosi-umkm/controller"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
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
	// page.Post("/register/admin", controller.RegisterAdmin)
	// page.Post("/admin/login", controller.LoginAdmin)
	// page.Post("/admin/logout", controller.Logout)

	// PRODUCT
	page.Post("/insert/product", controller.JWTAuth, controller.InsertDataProduct)       //menambahkan data product
	page.Get("/product", controller.GetAllProduct)                                       //menampilkan semua data product
	page.Get("/product/:id", controller.GetProductID)                                    //menampilkan data product berdasarkan id
	page.Put("/update/product/:id", controller.JWTAuth, controller.UpdateDataProduct)    //update data product
	page.Delete("/product/delete/:id", controller.JWTAuth, controller.DeleteProductByID) //delete data product
	// Menambahkan route dengan JWTAuth sebagai middleware
	page.Get("/product-seller", controller.JWTAuth, controller.GetProductsByUser)

	// CATEGORY
	page.Post("/insert/category", controller.InsertCategory)
	page.Get("/category", controller.GetAllCategory)
	page.Get("/category/:id", controller.GetCategoryByID)
	page.Put("/update/category/:id", controller.UpdateCategory)
	page.Delete("/category/delete/:id", controller.DeleteCategoryByID)

	// STORE

	page.Post("/insert/store", controller.InsertStore)
	page.Get("/store", controller.GetAllStores)
	page.Get("/store/:id", controller.GetStoreByID)
	page.Put("/update/store/:id", controller.UpdateStore)
	page.Delete("/store/delete/:id", controller.DeleteStoreByID)

	// Data Users

	page.Post("/insert/user", controller.InsertDataUser)
	page.Get("/users", controller.GetAllUsers)
	page.Get("/user/:id", controller.GetUserByID)
	page.Put("/update/user/:id", controller.UpdateDataUser)
	page.Delete("/user/delete/:id", controller.DeleteUserByID)

	// Status

	page.Post("/insert/status", controller.InsertStatus)
	page.Get("/statuses", controller.GetAllStatus)
	page.Get("/status/:id", controller.GetStatusByID)
	page.Put("/update/status/:id", controller.UpdateStatus)
	page.Delete("/status/delete/:id", controller.DeleteStatusByID)

	// page.Use(middleware.JWTMiddleware)
	page.Get("/dashboard", controller.DashboardPage)
	// Rute untuk menampilkan Swagger UI dan mendefinisikan URL untuk file swagger.yaml
	page.Get("/docs/*", swagger.HandlerDefault)

	// Pastikan rute untuk mengakses file swagger.yaml juga tersedia
	page.Get("/swagger.yaml", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.yaml")
	})

}
