package routers

import (
	"expense-tracker-api/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// Health check
	beego.Router("/api/v1/health", &controllers.HealthController{})

	// Namespace for Swagger-annotated routes
	ns := beego.NewNamespace("/api/v1",
		beego.NSNamespace("/auth",
			beego.NSRouter("/register", &controllers.AuthController{}, "post:Register"),
			beego.NSRouter("/login", &controllers.AuthController{}, "post:Login"),
		),
		beego.NSNamespace("/expenses",
			beego.NSRouter("/summary", &controllers.ExpenseController{}, "get:Summary"),
			beego.NSRouter("/", &controllers.ExpenseController{}, "post:Create;get:List"),
			beego.NSRouter("/:id", &controllers.ExpenseController{}, "get:GetOne;put:Update;delete:Delete"),
		),
	)
	beego.AddNamespace(ns)
}