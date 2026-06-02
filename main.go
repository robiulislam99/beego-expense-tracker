// main.go is the entry point of the expense-tracker-api application.
// It initializes the Beego framework, registers all routes, and starts the HTTP server.
//
// @title Expense Tracker API
// @version 1.0.0
// @description Personal Expense Tracker REST API built with Go and Beego v2
// @contact.email support@expense-tracker.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
package main

import (
	_ "expense-tracker-api/docs"
	_ "expense-tracker-api/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	// Run the Beego application.
	// All configuration is loaded from conf/app.conf
	beego.Run()
}