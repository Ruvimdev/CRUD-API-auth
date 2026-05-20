package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/swaggo/http-swagger"
	_ "github.com/lib/pq" // Тот самый драйвер с нижним подчеркиванием
	_ "golang-pgress/docs"

	"golang-pgress/internal/config"
	"golang-pgress/internal/handlers"
	"golang-pgress/internal/services"
	"golang-pgress/internal/storage"
)

// func loggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		log.Printf("получен запрос: %s %s", r.Method, r.URL.Path)
// 		next(w, r)
// 	}
// }


// @title Task Manager API
// @version 1.0
// @description API для управления задачами и категориями
// @host localhost:8080
// @BasePath /
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

	//создали канал
	taskChan := make(chan string, 10)
	//запустили воркера в фоне
	go services.EmailWorker(taskChan)

	//3 собираем архитектуру, dependency injection
	store := storage.NewStorage(db)
	service := services.NewAuthService(store, cfg)
	h := handlers.NewHandler(store, service, taskChan)
	

	//chi роутер
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST","PUT","PATCH","DELETE","OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.Logger)//встроенный логер
	r.Use(middleware.Recoverer)// защита от краша

	// ---ПУБЛИЧНЫЕ РОУТЫ (без защиты)----
	r.Post("/register", h.RegisterUser)
	r.Post("/login", h.LoginUser)
	r.Get("/swagger/*", httpSwagger.WrapHandler) //swagger ui

	//---- ПРИВАТНЫЕ РОУТЫ (под защитой) ----
	//r.Group создает изолированную группу, где работают свои правила
	r.Group(func(r chi.Router) {
	    //охранник на всю эту группу
		//передаем секретный ключ из конфига
		r.Use(handlers.AuthMiddleware(cfg.JWTSecret))

		//эти роуты требуют токен
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", h.GetTasks)
			r.Post("/", h.CreateTask)
			r.Delete("/{id}", h.DeleteTask)
			r.Patch("/{id}", h.UpdateTask)
		})

		r.Route("/categories", func(r chi.Router) {
			r.Post("/", h.CreateCategory)
			r.Get("/", h.GetCategory)
		})

	})



//GRACEFUL SHUTDOWN 

	srv := &http.Server{
		Addr: 	 ":8080",
		Handler: r,
	}

	//sigint = ctrl+c - os посылает процессу сигнал sigint(завершить процесс),
	//sigterm - деплой система останавливает контейнер этим сигналом 
	//канал получает уведомление ос 
	//канал - труба между горутинами, 1 = буффер на сигнал 
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) //пришел сигнал - клади в канал

	//сервер в горутине что бы не блокировать
	go func() {
		slog.Info("server launched", "port", "8080")
		err := srv.ListenAndServe() //программа висит тут 
		if err != nil && err != http.ErrServerClosed { //ErrServerClosed - штатное завершение работы(srv.shutdown())
			slog.Error("ошибка сервера", "error", err)
			os.Exit(1)
		}
	}()
	<-quit //блокируемся тут, код ниже будет выполнен после того как придет сигнал в канал
	slog.Info("останавливаем сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("ошибка при осатановке", "error", err)
	}
	slog.Info("сервер остановлен")
}


