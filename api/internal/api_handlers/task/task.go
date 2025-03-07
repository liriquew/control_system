package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time_manage/internal/api_handlers/auth"
	predictions_client "time_manage/internal/grpc/client"
	"time_manage/internal/storage"
)

type TaskAPI interface {
	CreateTask(w http.ResponseWriter, r *http.Request)
	GetTask(w http.ResponseWriter, r *http.Request)
	UpdateTask(w http.ResponseWriter, r *http.Request)
	DeleteTask(w http.ResponseWriter, r *http.Request)
	Predict(w http.ResponseWriter, r *http.Request)
}

type Task struct {
	infoLog    *log.Logger
	errorLog   *log.Logger
	storage    *storage.Storage
	taskClient *predictions_client.Client
}

func New(infoLog *log.Logger, errorLog *log.Logger, storage *storage.Storage, taskClient *predictions_client.Client) *Task {
	return &Task{
		infoLog:    infoLog,
		errorLog:   errorLog,
		storage:    storage,
		taskClient: taskClient,
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
	api.infoLog.Println("Create Task")

	var task storage.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		api.errorLog.Println("JSON error", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
	}
	uid := r.Context().Value(auth.UIDInterface{}).(int64)

	api.infoLog.Println(uid)

	taskID, err := api.storage.CreateTask(r.Context(), int64(uid), &task)
	if err != nil {
		api.errorLog.Println("Storage error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseID{ID: Int64String(taskID)})
}

func (api *Task) GetTask(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("Get Task")

	taskIDStr := r.URL.Query().Get("taskid")
	taskID, err := strconv.Atoi(taskIDStr)
	api.infoLog.Println(taskID, err)
	if err != nil {
		http.Error(w, "Некорректный taskID", http.StatusBadRequest)
		return
	}

	uid := r.Context().Value(auth.UIDInterface{}).(int64)
	api.infoLog.Println("UID:", uid)

	task, err := api.storage.GetTaskByID(r.Context(), int64(uid), int64(taskID))
	if err != nil {
		api.errorLog.Println("Storage error", err)
		if errors.Is(err, storage.ErrInvalidTaskID) {
			http.Error(w, "Некорректный taskID", http.StatusBadRequest)
			return
		} else if errors.Is(err, storage.ErrTaskNotFound) {
			http.Error(w, "Задача с таким taskID не найдена", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(task)
}

func (api *Task) UpdateTask(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("Update Task")

	taskIDStr := r.URL.Query().Get("taskid")
	taskID, err := strconv.Atoi(taskIDStr)
	api.infoLog.Println(taskID, err)
	if err != nil {
		http.Error(w, "invalid taskID", http.StatusBadRequest)
		return
	}

	var task storage.Task
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		api.errorLog.Println("JSON error", err)
		http.Error(w, "Bad json", http.StatusBadRequest)
		return
	}
	task.ID = int64(taskID)

	uid := r.Context().Value(auth.UIDInterface{}).(int64)
	if task.ActualTime != nil && *task.ActualTime != 0 {
		api.infoLog.Println("GRPC CALL")
		api.infoLog.Println(&task)
		err := api.taskClient.RecalculateAndSaveTask(r.Context(), uid, &task)
		if err != nil {
			if errors.Is(err, predictions_client.ErrInvalidArgument) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			api.errorLog.Println("gRPC error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		// перерасчет не требуется
		err = api.storage.UpdateTask(r.Context(), int64(uid), int64(taskID), &task)
		if err != nil {
			if errors.Is(err, storage.ErrInvalidTaskID) {
				http.Error(w, "Некорректный taskID", http.StatusBadRequest)
				return
			} else if errors.Is(err, storage.ErrTaskNotFound) {
				http.Error(w, "Задача с таким taskID не найдена", http.StatusBadRequest)
				return
			}
			api.errorLog.Println("Storage error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (api *Task) DeleteTask(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("Delete Task")

	taskIDStr := r.URL.Query().Get("taskid")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidTaskID) {
			http.Error(w, "invalid taskID", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uid := r.Context().Value(auth.UIDInterface{}).(int64)

	err = api.storage.DeleteTask(r.Context(), int64(uid), int64(taskID))
	if err != nil {
		if errors.Is(err, storage.ErrInvalidTaskID) {
			http.Error(w, "invalid taskID", http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = api.taskClient.Recalculate(r.Context(), uid)
	if err != nil && !errors.Is(err, predictions_client.ErrFailedPrecondition) {
		if errors.Is(err, predictions_client.ErrInvalidArgument) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		api.errorLog.Println("gRPC error:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (api *Task) Predict(w http.ResponseWriter, r *http.Request) {
	api.infoLog.Println("Predict")

	uid := r.Context().Value(auth.UIDInterface{}).(int64)
	plannedTimeStr := r.URL.Query().Get("planned_time")
	plannedTime, err := strconv.ParseFloat(plannedTimeStr, 64)
	api.infoLog.Println(plannedTime, err)
	if err != nil || plannedTime <= 0 {
		http.Error(w, "Некорректное планируемое время", http.StatusBadRequest)
		return
	}

	actualTime, err := api.taskClient.Predict(r.Context(), uid, plannedTime)
	if err != nil {
		if errors.Is(err, predictions_client.ErrFailedPrecondition) {
			http.Error(w, "Для того, чтобы сделать прогноз, хотя бы одна задача должна быть завершена", http.StatusBadRequest)
			return
		} else if errors.Is(err, predictions_client.ErrInvalidArgument) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		api.errorLog.Println("gRPC error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseActualTime{
		ActualTime: float64String(actualTime),
	})
}
