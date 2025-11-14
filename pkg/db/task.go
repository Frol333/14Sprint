package db

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

// Task хранит параметры задачи. Поле ID возвращается после добавления в БД.
type Task struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет задачу в таблицу scheduler и возвращает id новой записи.
func AddTask(ctx context.Context, task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	// Используем глобальную DB из пакета db
	res, err := DB.ExecContext(ctx, query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Tasks возвращает список задач с ограничением по количеству (limit).
// Задачи сортируются по дате ascending (по возрастанию).
func Tasks(limit int) ([]*Task, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := DB.Query(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {

		t := &Task{}
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// гарантия, что возвращаем не nil слайс
	if tasks == nil {
		tasks = make([]*Task, 0)
	}
	return tasks, nil

}

// GetTask возвращает задачу по ID (id в БД хранится как целое число, возвращаем как строку)
func GetTask(ctx context.Context, id int64) (*Task, error) {
	var t Task
	row := DB.QueryRowContext(ctx, `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id)
	if err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Задача не найдена")
		}
		return nil, err
	}

	return &t, nil
}

// UpdateTask обновляет запись задачи по её ID
func UpdateTask(ctx context.Context, task *Task) error {

	res, err := DB.ExecContext(ctx, `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("Задача не найдена")
	}
	return nil
}

// DeleteTask удаляет задачу по ID
func DeleteTask(ctx context.Context, id int64) error {
	res, err := DB.ExecContext(ctx, `DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("Задача не найдена")
	}
	return nil
}

// UpdateDate обновляет только колонку date у задачи
func UpdateDate(next string, id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id")
	}
	res, err := DB.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, next, idInt)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("Задача не найдена")
	}
	return nil
}
