// Package main содержит точку входа серверного приложения GophKeeper.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"github.com/victor/gophkeeper/internal/common/protocol"
	"github.com/victor/gophkeeper/internal/server/config"
	"github.com/victor/gophkeeper/internal/server/crypto"
	"github.com/victor/gophkeeper/internal/server/handlers"
	"github.com/victor/gophkeeper/internal/server/middleware"
	"github.com/victor/gophkeeper/internal/server/storage"
)

func main() {
	// Параметры командной строки
	dbPath := flag.String("db", "", "path to SQLite database file (overrides GOPHKEEPER_DB_PATH)")
	addr := flag.String("addr", "", "server address (overrides GOPHKEEPER_SERVER_ADDRESS)")
	flag.Parse()

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Переопределяем из флагов, если указаны (флаги имеют приоритет)
	if *dbPath != "" {
		cfg.Database.Path = *dbPath
	}
	if *addr != "" {
		cfg.Server.Address = *addr
	}

	// Инициализируем JWT конфигурацию
	crypto.SetJWTConfig(
		cfg.JWT.SecretKey,
		cfg.JWT.ExpirationTime,
		cfg.JWT.RefreshExpirationTime,
	)

	// Создаем хранилище
	st, err := storage.NewSQLiteStorage(cfg.Database.Path)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	// ВАЖНО: В production мастер-пароль должен быть только на клиенте!
	// Это временное решение. Получаем из переменной окружения или используем дефолт для dev.
	masterPassword := viper.GetString("master_password")
	if masterPassword == "" {
		masterPassword = "dev-master-password"
		log.Println("WARNING: Using default master password. Set GOPHKEEPER_MASTER_PASSWORD in production!")
	}

	// Создаем handlers
	authHandler := handlers.NewAuthHandler(st)
	dataHandler := handlers.NewDataHandler(st, masterPassword)

	// Настраиваем роутинг
	mux := http.NewServeMux()

	// Публичные endpoints (без аутентификации)
	mux.HandleFunc("POST "+protocol.APIPrefix+"/register", authHandler.Register)
	mux.HandleFunc("POST "+protocol.APIPrefix+"/login", authHandler.Login)
	mux.HandleFunc("POST "+protocol.APIPrefix+"/refresh", authHandler.Refresh)

	// Защищенные endpoints (с аутентификацией)
	authMiddleware := middleware.AuthMiddleware
	mux.HandleFunc("POST "+protocol.APIPrefix+"/logout", authMiddleware(http.HandlerFunc(authHandler.Logout)).ServeHTTP)
	mux.HandleFunc("GET "+protocol.APIPrefix+"/data", authMiddleware(http.HandlerFunc(dataHandler.GetAllData)).ServeHTTP)
	mux.HandleFunc("GET "+protocol.APIPrefix+"/data/{id}", authMiddleware(http.HandlerFunc(dataHandler.GetData)).ServeHTTP)
	mux.HandleFunc("POST "+protocol.APIPrefix+"/data", authMiddleware(http.HandlerFunc(dataHandler.CreateData)).ServeHTTP)
	mux.HandleFunc("PUT "+protocol.APIPrefix+"/data/{id}", authMiddleware(http.HandlerFunc(dataHandler.UpdateData)).ServeHTTP)
	mux.HandleFunc("DELETE "+protocol.APIPrefix+"/data/{id}", authMiddleware(http.HandlerFunc(dataHandler.DeleteData)).ServeHTTP)

	// Добавляем middleware для логирования
	handler := middleware.LoggingMiddleware(mux)

	// Создаем HTTP сервер с настройками из конфигурации
	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Server starting on %s", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Ожидаем сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Останавливаем сервер (не принимаем новые соединения, ждем завершения текущих)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Закрываем соединение с БД
	if err := st.Close(); err != nil {
		log.Printf("Error closing storage: %v", err)
	}

	log.Println("Server exited")
}
