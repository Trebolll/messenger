package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"messenger/internal/db"
	"messenger/internal/handler"
	"messenger/internal/middleware"
	"messenger/internal/repository"
	"messenger/internal/service"
	"messenger/internal/service/websocket"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer func(d *sql.DB) {
		if err := d.Close(); err != nil {
			log.Printf("Ошибка закрытия БД: %v", err)
		}
	}(database)

	applyMigrations(database)

	hub := websocket.NewHub()
	go hub.Run()

	storageService, err := service.NewStorageService(
		os.Getenv("MINIO_ENDPOINT"),
		os.Getenv("MINIO_ACCESS_KEY"),
		os.Getenv("MINIO_SECRET_KEY"),
		os.Getenv("MINIO_BUCKET"),
	)
	if err != nil {
		log.Fatalf("Ошибка инициализации хранилища: %v", err)
	}

	userRepository := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService, hub, storageService)

	chatRepository := repository.NewChatRepository(database)
	chatService := service.NewChatService(chatRepository, userRepository, hub)
	chatHandler := handler.NewChatHandler(chatService, storageService)

	messageRepository := repository.NewMessageRepository(database)
	messageService := service.NewMessageService(messageRepository, chatRepository, hub)
	messageHandler := handler.NewMessageHandler(messageService)

	ratingRepository := repository.NewRatingRepository(database)
	ratingService := service.NewRatingService(ratingRepository, chatRepository, hub)
	ratingHandler := handler.NewRatingHandler(ratingService)
	messageService.SetVoteEnricher(ratingRepository)

	aiService := service.NewAIService()
	aiHandler := handler.NewAIHandler(aiService)

	attachmentRepository := repository.NewAttachmentRepository(database)
	attachmentService := service.NewAttachmentService(attachmentRepository, storageService)
	attachmentHandler := handler.NewAttachmentHandler(attachmentService)

	wsHandler := handler.NewWebSocketHandler(hub, "your_secret_key")

	r := gin.Default()
	r.LoadHTMLGlob("web/*.html")
	r.Static("/web", "./web")

	r.GET("/", func(c *gin.Context) { c.HTML(200, "index.html", nil) })

	r.POST("/api/register", userHandler.Register)
	r.POST("/api/login", userHandler.Login)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware("your_secret_key"))
	{
		api.POST("/chats/private", chatHandler.CreatePrivateChat)
		api.POST("/chats/group", chatHandler.CreateGroupChat)
		api.POST("/messages", messageHandler.SendMessage)
		api.PUT("/messages/:message_id", messageHandler.EditMessage)
		api.DELETE("/messages/:message_id", messageHandler.DeleteMessage)
		api.GET("/chats/:chat_id/messages", messageHandler.GetMessages)
		api.GET("/chats", chatHandler.GetUserChats)
		api.PUT("/chats/:chat_id/avatar", chatHandler.UpdateGroupAvatar)
		api.PUT("/chats/:chat_id", chatHandler.UpdateGroupInfo)
		api.GET("/chats/:chat_id/members", chatHandler.GetGroupMembers)
		api.POST("/chats/:chat_id/members", chatHandler.AddChatMember)
		api.DELETE("/chats/:chat_id/members/:user_id", chatHandler.RemoveChatMember)
		api.GET("/users/search", userHandler.SearchUsers)
		api.PUT("/users/profile", userHandler.UpdateProfile)
		api.PUT("/users/avatar", userHandler.UpdateAvatar)
		api.PUT("/users/status", userHandler.UpdateStatus)
		api.POST("/chats/:chat_id/read", messageHandler.MarkAsRead)
		api.POST("/ai/suggest", aiHandler.Suggest)
		api.POST("/chats/:chat_id/attachments", attachmentHandler.Upload)
		api.POST("/messages/:message_id/vote", ratingHandler.Vote)
		api.GET("/users/:user_id/rating", ratingHandler.GetUserRating)
	}

	r.GET("/api/ws", wsHandler.HandleWebSocket)

	log.Printf("Server started at port 8080")
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

// dbExists проверяет что таблица users уже существует (старая БД)
func dbExists(database *sql.DB) bool {
	var exists bool
	database.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'users'
		)
	`).Scan(&exists)
	return exists
}

func applyMigrations(database *sql.DB) {
	// ШАГ 1: Подготавливаем schema_migrations ДО инициализации migrate
	// чтобы он правильно читал текущую версию при старте

	var smExists bool
	database.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)
	`).Scan(&smExists)

	if !smExists && dbExists(database) {
		// Старая БД без migrate — создаём schema_migrations и ставим версию 1
		log.Printf("migrate: старая БД без версионирования — инициализируем на версии 1")
		if _, err := database.Exec(`
			CREATE TABLE schema_migrations (version bigint NOT NULL, dirty boolean NOT NULL);
			INSERT INTO schema_migrations (version, dirty) VALUES (1, false);
		`); err != nil {
			log.Fatalf("migrate: не удалось создать schema_migrations: %v", err)
		}
	} else if smExists {
		// Сбрасываем dirty флаг если есть (от предыдущих неудачных запусков)
		var dirty bool
		var version int64
		database.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&version, &dirty)
		if dirty {
			log.Printf("migrate: сбрасываем dirty флаг на версии %d", version)
			database.Exec(`UPDATE schema_migrations SET dirty = false`)
		}
	}

	// ШАГ 2: Инициализируем migrate с уже корректной schema_migrations
	driver, err := migratepg.WithInstance(database, &migratepg.Config{})
	if err != nil {
		log.Fatalf("migrate: не удалось создать драйвер: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migration",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("migrate: не удалось инициализировать: %v", err)
	}

	// ШАГ 3: Применяем новые миграции
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate: ошибка применения миграций: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.Printf("migrate: не удалось получить версию: %v", err)
		return
	}
	if dirty {
		log.Fatalf("migrate: БД в грязном состоянии на версии %d", version)
	}
	log.Printf("migrate: БД на версии %d ✓", version)
}
