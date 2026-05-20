package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"golang-pgress/internal/models"
	"golang-pgress/internal/services"
	"golang-pgress/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
)

type Handler struct {
	databaseLayer   	storage.TaskStorage
	authService     	*services.AuthService
	validator       	*validator.Validate
	emailChan chan		string
}

func NewHandler(s storage.TaskStorage, auth *services.AuthService, taskChan chan string) *Handler  {
	return &Handler{
		databaseLayer: s,
		authService: auth,
		validator: validator.New(),
		emailChan: taskChan,
	}
}

func sendJson(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}	

func sendError(w http.ResponseWriter, status int, message string) {
	sendJson(w, status, map[string]string{"error": message})
}


// хендлеры стали методами (h *Handler) ----

//создаем задачу
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var UserTastText models.IncomingTask

	err := json.NewDecoder(r.Body).Decode(&UserTastText)
	if err != nil {
		sendError(w, http.StatusBadRequest, "неверный формат джсон")
		return 
	}
	

	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		sendError(w, http.StatusInternalServerError, "не удалось получить id пользователя")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	newID, err := h.databaseLayer.CreateTask(ctx, UserTastText.TaskText, userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "ошибка сохранения в базу")
		return
	}

	//worker
	select {
	case h.emailChan <- UserTastText.TaskText:
	default:
		slog.Warn("email queue full, skipping")
	}

	response := map[string]interface{} {
		"message": "задача успешно создана",
		"task_id": newID,
	}

	sendJson(w, http.StatusCreated, response)
}


func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	pageNumber := 1
	limit := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNumber = p 
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l ,err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := (pageNumber - 1) * limit

	
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "не авторизован")
		return
	}

	tasks, err := h.databaseLayer.GetAllTasks(ctx, userID, limit, offset)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "ошибка получении данных с бд ")
		return
	}

	response := map[string]interface{} {
		"message": "задачи успешно получены",
		"tasks": tasks,
	}

	sendJson(w, http.StatusOK, response)
}


func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	IdString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(IdString)

	if err != nil {
		sendError(w, http.StatusBadRequest, "ошибка конвертации айди")
		return
	}
	if id == 0 {
		sendError(w, http.StatusBadRequest, "неверный айди")
		return
	}


	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	//достаем айди чела из контекста (удаляем задачу которая есть у человека, нет - не проходит)
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "не авторизован")
		return
	}

	result, err := h.databaseLayer.DeleteTask(ctx, id, userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "ошибка удаления с базы")
		return
	}

	response := map[string]interface{} {
		"message": "успешно удалено",
		"task_id": result,
	}
	sendJson(w, http.StatusOK, response)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	
	IdString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(IdString)
	if err != nil {
		sendError(w, http.StatusBadRequest, "ошибка получения айди")
		return
	}

	var input models.UpdateTaskInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		sendError(w, http.StatusBadRequest, "ошибка декодирования")
	}
	if err := h.validator.Struct(input); err != nil {
		sendError(w, http.StatusBadRequest, "ошибка валидации данных")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "не авторизован")
		return
	}

	updatedStatus, err := h.databaseLayer.UpdateTask(ctx, input.Status, id, userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "ошибка обновления информации в базе")
		return
	}

	response := map[string]interface{}{
		"message": "задача успешно обновлена",
		"task_status": updatedStatus,
	}
	sendJson(w, http.StatusOK, response)

}




func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var UserCategoryName models.CategoryInput

	err := json.NewDecoder(r.Body).Decode(&UserCategoryName)
	if err != nil {
		sendError(w, http.StatusBadRequest, "ошибка чтения джсон")
		return
	}

	if err := h.validator.Struct(UserCategoryName); err != nil {
		sendError(w, http.StatusBadRequest, "ошибка валидации даннхы")
		return
	}

	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		sendError(w, http.StatusBadRequest, "ошибка получения айди")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	newID, err := h.databaseLayer.CreateCategory(ctx, UserCategoryName.CategoryName, userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "ошибка сохранения в базу")
		return
	}

	response := map[string]interface{} {
		"message": "ваша категория создана",
		"category_id": newID,
	}
	sendJson(w, http.StatusOK, response)
} 

// @Summary Регистрация пользователя
// @Description Создает нового пользователя в базе
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param input body models.RegisterInput true "данные для регистрации"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var UserRegisterData models.RegisterInput

	err := json.NewDecoder(r.Body).Decode(&UserRegisterData)
	if err != nil {
		sendError(w, http.StatusBadRequest, "ошибка чтения джсон")
		return
	}

	err = h.validator.Struct(UserRegisterData)
	if err != nil {
		sendError(w, http.StatusBadRequest, "неверный формат почты или маленькая длина пароля")
		return
	}

	userId, err := h.authService.RegisterUser(r.Context(), UserRegisterData.Email, UserRegisterData.Password)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "ошибка сохранения в базу")
		return
	}
	response := map[string]interface{}{
		"message": "пользователь успешно создан",
		"user_id": userId,
	}
	sendJson(w, http.StatusOK, response)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var UserLoginData models.RegisterInput
	
	err := json.NewDecoder(r.Body).Decode(&UserLoginData)
	if err != nil {
		sendError(w, http.StatusBadRequest, "ошибка получения данных с джсон")
		return
	}

	err = h.validator.Struct(UserLoginData)
	if err != nil {
		sendError(w, http.StatusBadRequest, "неверный почта или пароль")
		return
	}

	token, err := h.authService.LoginUser(r.Context(), UserLoginData.Email, UserLoginData.Password)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "неверная почта или пароль")
		return
	}

	response := map[string]interface{} {
		"message": "юзер найден в базе",
		"token": token,
	}
	sendJson(w, http.StatusOK, response)
}

func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "не удалось получить айди юзера")
		return
	}

	categories, err := h.databaseLayer.GetCategories(ctx, userID)
	if err != nil {
		sendError(w, http.StatusBadRequest, "ошибка получение категорий")
		return
	}

	if categories == nil {
		categories = []models.Category{}
	} 

	response := map[string]interface{} {
		"message": "получены категории пользователя",
		"categories": categories,
	}
	sendJson(w, http.StatusOK, response)
}