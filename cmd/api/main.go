package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Тот самый драйвер с нижним подчеркиванием

	"golang-pgress/internal/config"
	"golang-pgress/internal/handlers"
	"golang-pgress/internal/storage"
)

// func loggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		log.Printf("получен запрос: %s %s", r.Method, r.URL.Path)
// 		next(w, r)
// 	}
// }



func main() {
	//логгер
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	
	//енв
	if err := godotenv.Load(); err != nil {
		slog.Error("ошибка чтения .env файла", "error", err)
	}
	
	//конфиг
	cfg := config.LoadConfig()
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", 
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBPassword)


	//2 открываем соединенние 
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("ошибка инициализации", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		slog.Error("база не отвечает", "error", err)
		os.Exit(1)
	}
	slog.Info("успешно подключение")

	//3 собираем архитектуру, dependency injection
	store := storage.NewStorage(db)
	h := handlers.NewHandler(store)
	
	//chi роутер
	r := chi.NewRouter()

	r.Use(middleware.Logger)//встроенный логер
	r.Use(middleware.Recoverer)// защита от краша

	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", h.GetTasks)
		r.Post("/", h.CreateTask)
		r.Delete("/{id}", h.DeleteTask)
		r.Patch("/{id}", h.UpdateTask)

	})
	slog.Info("сервер запущен", "port", "8080")

	http.ListenAndServe(":8080", r)
}


