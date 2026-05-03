package storage

import (
	"context"
	"database/sql"
	"errors"
	"golang-pgress/internal/models"
)

type TaskStorage interface {
	CreateTask(ctx context.Context,title string) (int, error)
	GetAllTasks(ctx context.Context) ([]models.Task, error)
	DeleteTask(ctx context.Context, id int) (int, error)
	UpdateTask(ctx context.Context, status string, id int) (string, error)
}

//обернули пул соединений
type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateTask(ctx context.Context,title string) (int, error) {
	query := `INSERT INTO tasks (title, status) VALUES ($1, $2) RETURNING id`

	var newID int
	
	err := s.db.QueryRowContext(ctx, query, title, "active").Scan(&newID)
	if err != nil {
		return 0, err
	}

	return newID, nil
}


func (s *Storage) GetAllTasks(ctx context.Context) ([]models.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, status FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tasks []models.Task
	
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Status); err != nil {
			continue
		}

		tasks = append(tasks, t)
	}
	return tasks, nil
}


func (s *Storage) DeleteTask(ctx context.Context, id int) (int, error) {
	
	result, err := s.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rowsAffected == 0 {
		return 0, errors.New("задача с таким айди не найдена")
	}
	return id, nil
}

func (s *Storage) UpdateTask(ctx context.Context, status string, id int) (string, error) {
	query := `UPDATE tasks SET status = $1 WHERE id = $2`

	result, err := s.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return "", errors.New("ошибка обновления бд")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", errors.New("задача не найдена")// не прямая ошибка, если изменений нет err = nil = status code 200
													// кастомная ошибка
	}

	if rowsAffected == 0 {
		return "", errors.New("задача не найдена")
	}

	return status, nil
}