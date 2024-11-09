package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

var DB *sql.DB

// InitDB инициализирует подключение к базе данных и создает таблицу, если ее еще нет
func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS user_actions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		user_name TEXT,
		action TEXT,
		timestamp TEXT
	);
	`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	DB = db
	return db, nil
}

// LogUserAction записывает действие пользователя в базу данных с учетом часового пояса
func LogUserAction(userID int64, userName string, action string) {
	// Устанавливаем временную зону
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Ошибка загрузки часового пояса: %v", err)
		loc = time.UTC // если не удалось загрузить временную зону, используем UTC
	}

	// Получаем текущее время в нужной временной зоне
	currentTime := time.Now().In(loc).Format("2006-01-02 15:04:05")

	// Записываем действие в базу данных
	query := `INSERT INTO user_actions (user_id, user_name, action, timestamp) VALUES (?, ?, ?, ?)`
	_, err = DB.Exec(query, userID, userName, action, currentTime)
	if err != nil {
		log.Printf("Ошибка при логировании действия: %v", err)
	}
}
