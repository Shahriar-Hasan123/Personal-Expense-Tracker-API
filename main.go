package main

// @title Expense Tracker API
// @version 1.0
// @description A REST API for managing personal expenses
// @host localhost:8080
// @BasePath /api/v1

import (
	_ "expense-tracker-api/docs"
	_ "expense-tracker-api/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
		beego.BConfig.WebConfig.StaticDir["/docs"] = "docs"
	}
	beego.Run()
}
