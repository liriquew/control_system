package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time_manage/internal/api_handlers/auth"
	"time_manage/internal/models"
	service "time_manage/internal/service/tasks"

	"github.com/go-chi/chi/v5"
)

type TaskAPI interface {
	CreateTask(w http.ResponseWriter, r *http.Request)
	GetTask(w http.ResponseWriter, r *http.Request)
	GetTaskList(w http.ResponseWriter, r *http.Request)
	UpdateTask(w http.ResponseWriter, r *http.Request)
	DeleteTask(w http.ResponseWriter, r *http.Request)
	Predict(w http.ResponseWriter, r *http.Request)
}

type Task struct {
	infoLog     *log.Logger
	errorLog    *log.Logger
	taskService *service.TaskService
}

func New(infoLog *log.Logger, errorLog *log.Logger, taskService *service.TaskService) *Task {
	return &Task{
		infoLog:     infoLog,
		errorLog:    errorLog,
		taskService: taskService,
	}
}

const (
	headerContentType = "Content-Type"
	jsonContentType   = "application/json"
)

type Int64String int64
type float64String float64

func (i Int64String) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%d", i))
}

func (f float64String) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%.2f", f))
}

type ResponseID struct {
	ID Int64String `json:"id"`
}

type ResponseActualTime struct {
	ActualTime float64String `json:"actual_time"`
}

func (api *Task) CreateTask(w http.ResponseWriter, r *http.Request) {
	task, err := models.TaskModelFromJson(r.Body)
	if err != nil {
		api.errorLog.Println("JSON error", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	task.UserID = r.Context().Value(auth.UIDInterface{}).(int64)

	task, err = api.taskService.CreateTask(r.Context(), task)
	if err != nil {
		if errors.Is(err, service.ErrBadTaskTile) {
			http.Error(w, "invalid title", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrBadTaskDescription) {
			http.Error(w, "invalid description", http.StatusBadRequest)
			return
		}
		api.errorLog.Println("Service error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	api.infoLog.Println(task)
	json.NewEncoder(w).Encode(ResponseID{ID: Int64String(task.ID)})
}

func (api *Task) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "incorrect taskID", http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(auth.UIDInterface{}).(int64)

	task, err := api.taskService.GetTaskByID(r.Context(), userID, int64(taskID))
	if err != nil {
		if errors.Is(err, service.ErrInvalidTaskID) {
			http.Error(w, "incorrect taskID", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(task)
}

func (api *Task) GetTaskList(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	var err error

	offsetStr := r.URL.Query().Get("offset")
	var offsetVal int64
	if len(offsetStr) != 0 {
		offsetVal, err = strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
	} else {
		offsetVal = 0
	}

	limitStr := r.URL.Query().Get("limit")
	var limitVal int64
	if len(offsetStr) != 0 {
		limitVal, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
	} else {
		limitVal = 10
	}

	tasks, err := api.taskService.GetTaskListWithOffset(r.Context(), userID, limitVal, offsetVal)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "tasks not found", http.StatusNotFound)
			return
		}

		api.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(tasks)
}

func (api *Task) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid taskID", http.StatusBadRequest)
		return
	}

	task, err := models.TaskModelFromJson(r.Body)
	if err != nil {
		api.errorLog.Println("JSON error", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	task.ID, task.UserID = int64(taskID), userID

	_, err = api.taskService.UpdateTaskAndMarkModel(r.Context(), task)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTaskID) {
			http.Error(w, "invalid taskID", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrNothingToUpdate) {
			http.Error(w, "nothing to update (empty fields)", http.StatusBadRequest)
			return
		}
		if !errors.Is(err, service.ErrModelNotFound) {
			// thats ok error
			api.errorLog.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (api *Task) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("bad taskID param"))
		return
	}

	userID := r.Context().Value(auth.UIDInterface{}).(int64)

	_, err = api.taskService.DeleteTaskAndMarkModel(r.Context(), userID, int64(taskID))
	if err != nil && !errors.Is(err, service.ErrModelNotFound) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (api *Task) Predict(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("Predict")

	plannedTimeStr := r.URL.Query().Get("planned_time")
	plannedTime, err := strconv.ParseFloat(plannedTimeStr, 64)
	api.infoLog.Println(plannedTime, err)
	if err != nil || plannedTime <= 0 {
		http.Error(w, "Некорректное планируемое время", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(auth.UIDInterface{}).(int64)

	actualTime, err := api.taskService.PredictTaskCompletionTime(r.Context(), userID, plannedTime)
	if err != nil {
		if errors.Is(err, service.ErrNoCompletedTasks) {
			http.Error(w, "Для того, чтобы сделать прогноз, хотя бы одна задача должна быть завершена", http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseActualTime{
		ActualTime: float64String(actualTime),
	})
}
