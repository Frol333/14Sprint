package db

import (
	"context"
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// schema — создание таблицы Scheduler и индекса по полю date
const schema = `CREATE TABLE IF NOT EXISTS scheduler (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date CHAR(8) NOT NULL DEFAULT "",
  title VARCHAR(128),
  comment TEXT,
  repeat VARCHAR(128)
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
`

const userSchema = `CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    password_hash TEXT NOT NULL
);
`

// Init открывает базу данных и, при необходимости, создает таблицу и индекс.
func Init(dbFile string) error {
	install := false
	if _, err := os.Stat(dbFile); err != nil {
		if os.IsNotExist(err) {
			install = true
		} else {
			return err
		}
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	DB = db

	if install {
		if _, err := DB.Exec(schema); err != nil {
			return err
		}
		if _, err := DB.Exec(userSchema); err != nil {
			return err
		}
		if _, err := DB.ExecContext(context.Background(), `SELECT id FROM users WHERE password_hash`); err != nil {
			return err
		}

	}
	return nil

}
