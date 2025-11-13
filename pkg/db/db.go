package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// schema содержит создание таблицы scheduler и индекса по date.
// date хранится как CHAR(8) (YYYYMMDD). title — VARCHAR(128), comment — TEXT, repeat — VARCHAR(128).
const schema = `CREATE TABLE IF NOT EXISTS scheduler (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date CHAR(8) NOT NULL DEFAULT "",
  title VARCHAR(128) NOT NULL,
  comment TEXT,
  repeat VARCHAR(128)
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init открывает базу данных по файлу dbFile.
// Теперь схема выполняется каждый раз, чтобы гарантировать существование таблицы.
func Init(dbFile string) error {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	DB = db

	// Гарантируем существование таблицы и индекса независимо от существования файла
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	return nil

}

// Close закрывает соединение с БД
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil

}
