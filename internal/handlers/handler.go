package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"golang-pgress/internal/models"
	"golang-pgress/internal/storage"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	databaseLayer storage.TaskStorage
}

func NewHandler(s storage.TaskStorage)	*Handler  {
	return &Handler{databaseLayer: s}
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
	
	if UserTastText.TaskText == "" {
		sendError(w, http.StatusBadRequest, "текст задачи пустой")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	
	defer cancel()

	newID, err := h.databaseLayer.CreateTask(ctx, UserTastText.TaskText)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "ошибка сохранения в базу")
		return
	}

	//горутина
	go func (taskText string)  {
		// bgCtx := context.Background()
		
		time.Sleep(5 * time.Second)

		println("ФОНОВАЯ ЗАДАЧА: Уведомление о задаче '" + taskText + "' успешно отправлено!")
	}(UserTastText.TaskText)

	response := map[string]interface{} {
		"message": "задача успешно создана",
		"task_id": newID,
	}

	sendJson(w, http.StatusCreated, response)
}


func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	
	defer cancel()

	tasks, err := h.databaseLayer.GetAllTasks(ctx)
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

	result, err := h.databaseLayer.DeleteTask(ctx, id)
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
		sendError(w, http.StatusBadRequest, "ошибка получения текста с джсон")
		return
	}

	if id == 0 {
		sendError(w, http.StatusBadRequest, "неверный айди")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	updatedStatus, err := h.databaseLayer.UpdateTask(ctx, input.Status, id)
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