package main

import (
	"fmt"
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
	"strings"
)

// Initialize инициализирует все необходимые подключения и зависимости
func Initialize() {
	var initErrors []string
	err := cache.Connect()
	if err != nil {
		initErrors = append(initErrors, err.Error())
	}

	err = db.Connect()
	if err != nil {
		initErrors = append(initErrors, err.Error())
	}

	err = logs.Connect()
	if err != nil {
		initErrors = append(initErrors, err.Error())
	}

	err = nats.Connect()
	if err != nil {
		initErrors = append(initErrors, err.Error())
	}

	if len(initErrors) != 0 {
		panic(fmt.Sprintf("Запуск приложения невозможен из-за следующих ошибок инициализации %s", strings.Join(initErrors, ",\n")))
	}

	go func() { // работа без логирования возможна, но на эту ошибку нужно будет обратить внимание
		err := nats.Subscribe()
		if err != nil {
			log.Printf("Получение логов невозможно, Nats возвращает в подписчике: %s\n", err.Error())
		}
	}()
}

func main() {
	if os.Getenv("APP_ENV") == "local" {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	Initialize()

	app := gin.Default()

	items := app.Group("/items")
	service := app.Group("/service")

	items.POST("/create", controllers.CreateItem)
	items.PATCH("/update", controllers.UpdateItem)
	items.DELETE("/remove", controllers.DeleteItem)
	items.GET("/list", controllers.GetItems)

	service.GET("/logs", controllers.GetLogs)

	err := app.Run(os.Getenv("APP_PORT"))
	if err != nil {
		panic(err)
	}
}
