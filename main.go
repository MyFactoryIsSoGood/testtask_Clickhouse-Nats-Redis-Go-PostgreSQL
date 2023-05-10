package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"hezzl/cache"
	"hezzl/controllers"
	"hezzl/db"
	"hezzl/logs"
	"hezzl/nats"
	"log"
	"os"
)

func main() {
	if os.Getenv("APP_ENV") != "docker" {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	cache.Connect()
	db.Connect()
	logs.Connect()
	nats.Connect()
	go nats.Subscribe()
	app := gin.Default()

	app.POST("/item/create", controllers.CreateItem)
	app.PATCH("/item/update", controllers.UpdateItem)
	app.DELETE("/item/remove", controllers.DeleteItem)
	app.GET("/items/list", controllers.GetItems)

	_ = app.Run(os.Getenv("APP_PORT"))
}
