package handlers

import (
	"context"
	"golang-pgress/internal/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- 1. Mock , структура притворяется бд, все методы интерфейса storage.TaskStorage
type MockStorage struct {}

func (m *MockStorage) CreateTask(ctx context.Context, title string) (int, error){
	return 1, nil
}

func (m *MockStorage) GetAllTasks(ctx context.Context) ([]models.Task, error){
	return []models.Task{}, nil
}

func (m *MockStorage) DeleteTask(ctx context.Context, id int) (int, error){
	return id, nil
}

func (m *MockStorage) UpdateTask(ctx context.Context, status string, id int) (string, error){
	return status, nil
}

// ---- 2 тест ---- 
func TestCreateTask_Success(t *testing.T) {
	//handler с фейк бд
	mock := &MockStorage{}
	h := NewHandler(mock) 

	//fake запрос
	body := strings.NewReader(`{"task_text": "купить хлеб"}`)
	r := httptest.NewRequest(http.MethodPost, "/tasks", body)

	//fake w ResponseWriter - запоминает ответ
	w := httptest.NewRecorder()

	h.CreateTask(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("ожидали 201, получили %d", w.Code)
	}
}

func TestGetTask_Success(t *testing.T) {
	mock := &MockStorage{}
	h := NewHandler(mock)

	r := httptest.NewRequest(http.MethodGet, "/tasks", nil)

	w := httptest.NewRecorder()

	h.GetTasks(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("ожидали 200, получили %d", w.Code)
	}
}

func TestCreateTask_EmptyText(t *testing.T) {
	mock := &MockStorage{}
	h := NewHandler(mock)

	body := strings.NewReader(`{"task_text": ""}`)
	r := httptest.NewRequest(http.MethodPost, "/tasks", body)

	w := httptest.NewRecorder()

	h.CreateTask(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидали 400 а получили %d", w.Code)
	}
}