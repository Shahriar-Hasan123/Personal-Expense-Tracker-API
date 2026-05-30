package routers

import (
	"expense-tracker-api/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// Health check
	beego.Router("/api/v1/health", &controllers.HealthController{})

	// Auth routes
	beego.Router("/api/v1/auth/register", &controllers.AuthController{}, "post:Register")
	beego.Router("/api/v1/auth/login", &controllers.AuthController{}, "post:Login")

	// Expense routes — summary must be before :id to avoid route conflict
	beego.Router("/api/v1/expenses/summary", &controllers.ExpenseController{}, "get:Summary")
	beego.Router("/api/v1/expenses", &controllers.ExpenseController{}, "post:Create;get:List")
	beego.Router("/api/v1/expenses/:id", &controllers.ExpenseController{}, "get:GetOne;put:Update;delete:Delete")
}