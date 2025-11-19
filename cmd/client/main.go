// Package main содержит точку входа клиентского приложения GophKeeper.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/victor/gophkeeper/internal/client/api"
	"github.com/victor/gophkeeper/internal/client/crypto"
	"github.com/victor/gophkeeper/internal/client/storage"
	"github.com/victor/gophkeeper/internal/common/models"
	"github.com/victor/gophkeeper/pkg/version"
)

func main() {
	// Параметры командной строки
	serverURL := flag.String("server", "http://localhost:8080", "server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	// Обработка команды version
	if command == "version" {
		fmt.Println(version.GetInfo())
		os.Exit(0)
	}

	// Инициализируем локальное хранилище
	homeDir, err := getHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}

	dbPath := filepath.Join(homeDir, ".gophkeeper", "client.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		log.Fatalf("failed to create config directory: %v", err)
	}

	localStorage, err := storage.NewLocalStorage(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize local storage: %v", err)
	}
	defer localStorage.Close()

	// Создаем API клиент
	client := api.NewClient(*serverURL)

	// Загружаем сохраненные токены
	token, refreshToken, err := localStorage.GetTokens(context.Background())
	if err == nil {
		client.SetToken(token)
		client.SetRefreshToken(refreshToken)
	}

	ctx := context.Background()

	// Обработка команд
	switch command {
	case "register":
		if len(args) < 3 {
			log.Fatal("usage: gophkeeper register <login> <password>")
		}
		handleRegister(ctx, client, localStorage, args[1], args[2])

	case "login":
		if len(args) < 3 {
			log.Fatal("usage: gophkeeper login <login> <password>")
		}
		handleLogin(ctx, client, localStorage, args[1], args[2])

	case "logout":
		handleLogout(ctx, client, localStorage)

	case "list":
		handleList(ctx, client, localStorage)

	case "get":
		if len(args) < 2 {
			log.Fatal("usage: gophkeeper get <id>")
		}
		handleGet(ctx, client, localStorage, args[1])

	case "add":
		if len(args) < 4 {
			log.Fatal("usage: gophkeeper add <type> <name> <data> [metadata]")
		}
		metadata := ""
		if len(args) >= 5 {
			metadata = args[4]
		}
		handleAdd(ctx, client, localStorage, args[1], args[2], args[3], metadata)

	case "update":
		if len(args) < 3 {
			log.Fatal("usage: gophkeeper update <id> <data>")
		}
		handleUpdate(ctx, client, localStorage, args[1], args[2])

	case "delete":
		if len(args) < 2 {
			log.Fatal("usage: gophkeeper delete <id>")
		}
		handleDelete(ctx, client, localStorage, args[1])

	case "sync":
		handleSync(ctx, client, localStorage)

	default:
		fmt.Printf("unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("GophKeeper CLI Client")
	fmt.Println("\nCommands:")
	fmt.Println("  register <login> <password>  - Register new user")
	fmt.Println("  login <login> <password>     - Login user")
	fmt.Println("  logout                       - Logout user")
	fmt.Println("  list                         - List all data records")
	fmt.Println("  get <id>                     - Get data record by ID")
	fmt.Println("  add <type> <name> <data> [metadata] - Add new data record")
	fmt.Println("  update <id> <data>           - Update data record")
	fmt.Println("  delete <id>                  - Delete data record")
	fmt.Println("  sync                         - Sync with server")
	fmt.Println("  version                      - Show version and build date")
	fmt.Println("\nData types: CREDENTIALS, TEXT, BINARY, CARD")
}

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func handleRegister(ctx context.Context, client *api.Client, localStorage *storage.LocalStorage, login, password string) {
	resp, err := client.Register(ctx, login, password)
	if err != nil {
		log.Fatalf("registration failed: %v", err)
	}

	if err := localStorage.SaveTokens(ctx, resp.Token, resp.RefreshToken); err != nil {
		log.Fatalf("failed to save tokens: %v", err)
	}

	fmt.Println("Registration successful!")
}

func handleLogin(ctx context.Context, client *api.Client, localStorage *storage.LocalStorage, login, password string) {
	resp, err := client.Login(ctx, login, password)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	if err := localStorage.SaveTokens(ctx, resp.Token, resp.RefreshToken); err != nil {
		log.Fatalf("failed to save tokens: %v", err)
	}

	fmt.Println("Login successful!")
}

func handleLogout(ctx context.Context, client *api.Client, localStorage *storage.LocalStorage) {
	if err := client.Logout(ctx); err != nil {
		log.Fatalf("logout failed: %v", err)
	}

	// Удаляем токены из локального хранилища
	if _, _, err := localStorage.GetTokens(ctx); err == nil {
		// Токены есть, удаляем их
		localStorage.SaveTokens(ctx, "", "")
	}

	fmt.Println("Logout successful!")
}

func handleList(ctx context.Context, client *api.Client, localStorage *storage.LocalStorage) {
	records, err := client.GetAllData(ctx)
	if err != nil {
		log.Fatalf("failed to get data: %v", err)
	}

	if len(records) == 0 {
		fmt.Println("No records found")
		return
	}

	fmt.Println("Data records:")
	for _, record := range records {
		fmt.Printf("  ID: %s, Type: %s, Name: %s\n", record.ID, record.Type, record.Name)
	}
}

func handleGet(ctx context.Context, client *api.Client, localStorage *storage.LocalStorage, recordID string) {
	record, err := client.GetData(ctx, recordID)
	if err != nil {
		log.Fatalf("failed to get record: %v", err)
	}

	// Расшифровываем данные
	masterPassword := "default-master-password" // TODO: запрашивать у пользователя
	decryptedData, err := crypto.DecryptData([]byte(record.Data), masterPassword)
	if err != nil {
		log.Fatalf("failed to decrypt data: %v", err)
	}

	fmt.Printf("ID: %s\n", record.ID)
	fmt.Printf("Type: %s\n", record.Type)
	fmt.Printf("Name: %s\n", record.Name)
	fmt.Printf("Metadata: %s\n", record.Metadata)
	fmt.Printf("Data: %s\n", string(decryptedData))
	fmt.Printf("Version: %d\n", record.Version)
}

func handleAdd(ctx context.Context, client *api.Client, localStorage *storage.LocalStorage, dataType, name, data, metadata string) {
	dataTypeEnum := models.DataType(strings.ToUpper(dataType))
	if !dataTypeEnum.IsValid() {
		log.Fatalf("invalid data type: %s", dataType)
	}

	req := &models.CreateDataRequest{
		Type:     dataTypeEnum,
		Name:     name,
		Data:     data,
		Metadata: metadata,
	}

	record, err := client.CreateData(ctx, req)
	if err != nil {
		log.Fatalf("failed to create record: %v", err)
	}

	// Сохраняем локально
	if err := localStorage.SaveDataRecord(ctx, record); err != nil {
		log.Printf("warning: failed to save locally: %v", err)
	}

	fmt.Printf("Record created: %s\n", record.ID)
}

func handleUpdate(ctx context.Context, client *api.Client, localStorage *storage.LocalStorage, recordID, data string) {
	// Получаем текущую запись для версии
	record, err := client.GetData(ctx, recordID)
	if err != nil {
		log.Fatalf("failed to get record: %v", err)
	}

	req := &models.UpdateDataRequest{
		Data:    data,
		Version: record.Version,
	}

	updated, err := client.UpdateData(ctx, recordID, req)
	if err != nil {
		log.Fatalf("failed to update record: %v", err)
	}

	// Сохраняем локально
	if err := localStorage.SaveDataRecord(ctx, updated); err != nil {
		log.Printf("warning: failed to save locally: %v", err)
	}

	fmt.Printf("Record updated: %s\n", updated.ID)
}

func handleDelete(ctx context.Context, client *api.Client, localStorage *storage.LocalStorage, recordID string) {
	if err := client.DeleteData(ctx, recordID); err != nil {
		log.Fatalf("failed to delete record: %v", err)
	}

	// Удаляем локально
	if err := localStorage.DeleteDataRecord(ctx, recordID); err != nil {
		log.Printf("warning: failed to delete locally: %v", err)
	}

	fmt.Printf("Record deleted: %s\n", recordID)
}

func handleSync(ctx context.Context, client *api.Client, localStorage *storage.LocalStorage) {
	// Получаем данные с сервера
	serverRecords, err := client.GetAllData(ctx)
	if err != nil {
		log.Fatalf("failed to sync: %v", err)
	}

	// Сохраняем все записи локально
	for _, record := range serverRecords {
		if err := localStorage.SaveDataRecord(ctx, record); err != nil {
			log.Printf("warning: failed to save record %s: %v", record.ID, err)
		}
	}

	fmt.Printf("Synced %d records\n", len(serverRecords))
}
