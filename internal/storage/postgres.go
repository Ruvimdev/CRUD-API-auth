package storage

import (
	"context"
	"database/sql"
	"errors"
	"golang-pgress/internal/models"
)

type TaskStorage interface {
	CreateTask(ctx context.Context, title string, userID int) (int, error)
	GetAllTasks(ctx context.Context, userID int) ([]models.Task, error)
	DeleteTask(ctx context.Context, id int, userID int) (int, error)
	UpdateTask(ctx context.Context, status string, id int, userID int) (string, error)
    CreateUser(ctx context.Context, email string, passwordHash string) (int, error) 
    GetUserByEmail(ctx context.Context, email string) (models.User, error)
	CreateCategory(ctx context.Context, name string, userID int) (int, error)
	GetCategories(ctx context.Context, userID int) ([]models.Category, error)
}

//обернули пул соединений
type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

//tasks
//--create
func (s *Storage) CreateTask(ctx context.Context, title string, userID int) (int, error) {
	query := `INSERT INTO tasks (title, status, user_id) VALUES ($1, $2, $3) RETURNING id`

	var newID int
	
	err := s.db.QueryRowContext(ctx, query, title, "active", userID).Scan(&newID)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

//get
func (s *Storage) GetAllTasks(ctx context.Context, userID int) ([]models.Task, error) {
	
	query := `SELECT id, title, status FROM tasks WHERE user_id = $1 ORDER BY id`
	rows, err := s.db.QueryContext(ctx, query, userID)
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

//delete
func (s *Storage) DeleteTask(ctx context.Context, id int, userID int) (int, error) {
	
	query := `DELETE FROM tasks WHERE id = $1 AND user_Id = $2`
	result, err := s.db.ExecContext(ctx, query, id, userID)
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

//update
func (s *Storage) UpdateTask(ctx context.Context, status string, id int, userID int) (string, error) {
	query := `UPDATE tasks SET status = $1 WHERE id = $2 AND user_id = $3`

	result, err := s.db.ExecContext(ctx, query, status, id, userID)
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

//users
//-create
func (s *Storage) CreateUser(ctx context.Context, email string, passwordHash string) (int, error) {
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`
	var newId int
	err := s.db.QueryRowContext(ctx, query, email, passwordHash).Scan(&newId)
	if err != nil {
		return 0, err
	}
	return newId, nil
}

//-get
func (s *Storage) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	query := `SELECT id, email, password_hash FROM users WHERE email = $1`

	err := s.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.Passwordhash)
	if err != nil {
		return u, err
	}
	return u, nil
}

//categories
//-create
func (s *Storage) CreateCategory(ctx context.Context, name string, userID int) (int, error) {
	query := `INSERT INTO categories (name, user_id) VALUES ($1, $2) RETURNING id`

	var newID int 

	err := s.db.QueryRowContext(ctx, query, name, userID).Scan(&newID)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (s *Storage) GetCategories(ctx context.Context, userID int) ([]models.Category, error) {
	query := `SELECT id, name FROM categories WHERE user_id = $1 ORDER BY id`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var categories []models.Category

	for rows.Next() {
		var c models.Category

		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			continue
		}
		categories = append(categories, c)
	}
	return categories, nil
}