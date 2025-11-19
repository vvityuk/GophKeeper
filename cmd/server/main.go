// Package main содержит точку входа серверного приложения GophKeeper.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/victor/gophkeeper/internal/common/protocol"
	"github.com/victor/gophkeeper/internal/server/handlers"
	"github.com/victor/gophkeeper/internal/server/middleware"
	"github.com/victor/gophkeeper/internal/server/storage"
)

func main() {
	// Параметры командной строки
	dbPath := flag.String("db", "gophkeeper.db", "path to SQLite database file")
	addr := flag.String("addr", ":8080", "server address")
	flag.Parse()

	// Создаем хранилище
	st, err := storage.NewSQLiteStorage(*dbPath)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	defer st.Close()

	// Создаем handlers
	authHandler := handlers.NewAuthHandler(st)
	dataHandler := handlers.NewDataHandler(st)

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

	// Запускаем сервер
	log.Printf("Server starting on %s", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
		os.Exit(1)
	}
}
