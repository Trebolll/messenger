package main

import (
	"log"
	"messenger/internal/db"
	"messenger/internal/handler"
	"messenger/internal/repository"
	"messenger/internal/service"

	"database/sql"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	database, err := db.InitDB()

	if err != nil {
		panic(err)
	}
	defer func(database *sql.DB) {
		err := database.Close()
		if err != nil {
			panic(err)
		}
	}(database)

	applyMigrations(database)

	userRepository := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/register", userHandler.Register)
	}

	log.Printf("Server started at port 8080")
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func applyMigrations(db *sql.DB) {
	// Читаем SQL файл
	query, err := os.ReadFile("internal/db/migration/001_init.sql")
	if err != nil {
		log.Fatalf("Ошибка чтения файла миграции: %v", err)
	}

	_, err = db.Exec(string(query))
	if err != nil {
		log.Fatalf("Ошибка применения миграции: %v", err)
	}

	log.Println("Миграции успешно применены!")
}
