// Package routers registers all API routes for the expense-tracker-api application.
// All routes are prefixed with /api/v1 following RESTful conventions.
package routers

import (
	"expense-tracker-api/controllers"

	beego "github.com/beego/beego/v2/server/web"
	httpSwagger "github.com/swaggo/http-swagger"
)

// init registers all application routes when the package is loaded.
func init() {
	// Swagger UI endpoint
	beego.Handler("/swagger/*", httpSwagger.WrapHandler)

	// Health check endpoint
	beego.Router("/api/v1/health", &controllers.AuthController{}, "get:HealthCheck")

	// Auth endpoints
	beego.Router("/api/v1/auth/register", &controllers.AuthController{}, "post:Register")
	beego.Router("/api/v1/auth/login", &controllers.AuthController{}, "post:Login")

	// Expense endpoints
	beego.Router("/api/v1/expenses", &controllers.ExpenseController{}, "post:Create")
	beego.Router("/api/v1/expenses", &controllers.ExpenseController{}, "get:List")

	// Summary must be registered BEFORE /:id to avoid conflict
	beego.Router("/api/v1/expenses/summary", &controllers.ExpenseController{}, "get:Summary")

	beego.Router("/api/v1/expenses/:id", &controllers.ExpenseController{}, "get:GetOne")
	beego.Router("/api/v1/expenses/:id", &controllers.ExpenseController{}, "put:Update")
	beego.Router("/api/v1/expenses/:id", &controllers.ExpenseController{}, "delete:Delete")
}