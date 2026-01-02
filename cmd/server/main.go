package main

import (
	"log"
	"messenger/internal/db"
	"messenger/internal/handler"
	"messenger/internal/middleware"
	"messenger/internal/repository"
	"messenger/internal/service"
	"messenger/internal/service/websocket"

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

	hub := websocket.NewHub()
	go hub.Run()

	userRepository := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	chatRepository := repository.NewChatRepository(database)
	chatService := service.NewChatService(chatRepository, userRepository, hub)
	chatHandler := handler.NewChatHandler(chatService)

	messageRepository := repository.NewMessageRepository(database)
	messageService := service.NewMessageService(messageRepository, chatRepository, hub)
	messageHandler := handler.NewMessageHandler(messageService)

	wsHandler := handler.NewWebSocketHandler(hub, "your_secret_key")

	r := gin.Default()

	r.POST("/api/register", userHandler.Register)
	r.POST("/api/login", userHandler.Login)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware("your_secret_key"))
	{
		api.POST("/chats/private", chatHandler.CreatePrivateChat)
		api.POST("/chats/group", chatHandler.CreateGroupChat)
		api.POST("/messages", messageHandler.SendMessage)
		api.GET("/chats/:chat_id/messages", messageHandler.GetMessages)
		api.GET("/ws", wsHandler.HandleWebSocket)
		api.GET("/chats", chatHandler.GetUserChats)
		api.GET("/users/search", userHandler.SearchUsers)
		api.POST("/chats/:chat_id/read", messageHandler.MarkAsRead)
	}

	log.Printf("Server started at port 8080")
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func applyMigrations(db *sql.DB) {
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
