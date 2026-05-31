// main.go is the entry point of the expense-tracker-api application.
// It initializes the Beego framework, registers all routes, and starts the HTTP server.
package main

import (
	_ "expense-tracker-api/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	// Run the Beego application.
	// All configuration is loaded from conf/app.conf
	beego.Run()
}