package task

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/liriquew/control_system/internal/api_handlers/auth"
	"github.com/liriquew/control_system/internal/api_handlers/groups"
	groupsclient "github.com/liriquew/control_system/internal/grpc/clients/groups"
	predictionsclient "github.com/liriquew/control_system/internal/grpc/clients/predictions"
	tasksclient "github.com/liriquew/control_system/internal/grpc/clients/tasks"
	jsontools "github.com/liriquew/control_system/internal/lib/json_tools"
	"github.com/liriquew/control_system/internal/models"
	"github.com/liriquew/control_system/pkg/logger/sl"
)

type TaskAPI interface {
	CreateTask(w http.ResponseWriter, r *http.Request)
	GetTask(w http.ResponseWriter, r *http.Request)
	GetTaskList(w http.ResponseWriter, r *http.Request)
	UpdateTask(w http.ResponseWriter, r *http.Request)
	DeleteTask(w http.ResponseWriter, r *http.Request)
	PredictTask(w http.ResponseWriter, r *http.Request)
	GetTags(w http.ResponseWriter, r *http.Request)
	PredictTags(w http.ResponseWriter, r *http.Request)
	PredictUncreatedTask(w http.ResponseWriter, r *http.Request)
}

type Tasks struct {
	log               *slog.Logger
	taskClient        tasksclient.TasksClient
	groupsClient      groupsclient.GroupsClient
	predictionsClient predictionsclient.PredictionsClient
}

func NewTasksService(
	log *slog.Logger,
	taskClient tasksclient.TasksClient,
	groupsClient groupsclient.GroupsClient,
	predictionsClient predictionsclient.PredictionsClient,
) *Tasks {
	return &Tasks{
		log:               log,
		taskClient:        taskClient,
		groupsClient:      groupsClient,
		predictionsClient: predictionsClient,
	}
}

func (t *Tasks) CreateTask(w http.ResponseWriter, r *http.Request) {
	task, err := models.TaskModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	t.log.Debug("recieved task to create", slog.Any("task", task))

	if task.PlannedTime <= 0 {
		http.Error(w, "plannedTime must be greated than zero", http.StatusBadRequest)
		return
	}
	if task.ActualTime < 0 {
		http.Error(w, "actualTime must be greated than zero", http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		http.Error(w, "empty title", http.StatusBadRequest)
		return
	}
	if task.Description == "" {
		http.Error(w, "empty description", http.StatusBadRequest)
		return
	}

	if task.AssignedTo != 0 {
		groupID := groups.GetGroupID(r)
		if err := t.groupsClient.CheckMemberPermission(r.Context(), task.AssignedTo, groupID); err != nil {
			if errors.Is(err, groupsclient.ErrPermissionDenied) {
				http.Error(w, "assigned_to user not in group", http.StatusBadRequest)
				return
			}

			t.log.Error("error while checking assigned_to user group permission", slog.Any("task", task), slog.Int64("groupID", groupID), sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	task.GroupID = groups.GetGroupID(r)

	taskID, err := t.taskClient.CreateTask(r.Context(), task)
	if err != nil {
		if errors.Is(err, tasksclient.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, tasksclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, tasksclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		t.log.Error("error while creating task", slog.Any("task", task), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	jsontools.WriteInt64ID(w, taskID)
}

func (t *Tasks) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := GetTaskID(r)

	task, err := t.taskClient.GetTask(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, tasksclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, tasksclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		t.log.Error("error while getting task", slog.Int64("taskID", taskID), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.log.Debug("task", slog.Any("task", task))

	jsontools.WtiteJSON(w, task)
}

func (t *Tasks) GetTaskList(w http.ResponseWriter, r *http.Request) {
	offset := GetOffset(r)
	tasks, err := t.taskClient.GetTaskList(r.Context(), offset)
	if err != nil {
		t.log.Error("error while getting task list", slog.Int64("offset", offset), sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jsontools.WtiteJSON(w, tasks)
}

func (t *Tasks) UpdateTask(w http.ResponseWriter, r *http.Request) {
	task, err := models.TaskModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if task.Title == "" && task.Description == "" && task.PlannedTime == 0.0 && task.ActualTime == 0 && task.AssignedTo == 0 && len(task.Tags) == 0 {
		http.Error(w, "nothing to update", http.StatusBadRequest)
		return
	}
	if task.PlannedTime < 0 {
		http.Error(w, "plannedTime must be greated than zero", http.StatusBadRequest)
		return
	}
	if groupID := groups.GetGroupID(r); task.AssignedTo != 0 && groupID == 0 {
		http.Error(w, "groupID required to assign user", http.StatusBadRequest)
		return
	}
	task.ID = GetTaskID(r)
	task.GroupID = groups.GetGroupID(r)

	if task.AssignedTo != 0 {
		groupID := groups.GetGroupID(r)
		if err := t.groupsClient.CheckMemberPermission(r.Context(), task.AssignedTo, groupID); err != nil {
			if errors.Is(err, groupsclient.ErrPermissionDenied) {
				http.Error(w, "assigned_to user not in group", http.StatusBadRequest)
				return
			}

			t.log.Error("error while checking assigned_to user group permission", slog.Any("task", task), slog.Int64("groupID", groupID), sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err = t.taskClient.UpdateTask(r.Context(), task)
	if err != nil {
		if errors.Is(err, tasksclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, tasksclient.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, tasksclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		t.log.Error("error while updating task", slog.Any("task", task), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Tasks) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := GetTaskID(r)

	err := t.taskClient.DeleteTask(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, tasksclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, tasksclient.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, tasksclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		}

		t.log.Error("error while deleting task", slog.Int64("taskID", taskID), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Tasks) PredictTask(w http.ResponseWriter, r *http.Request) {
	taskID := GetTaskID(r)

	predictedTask, err := t.taskClient.PredictTask(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, tasksclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, tasksclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, tasksclient.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		t.log.Error("error while predicting task", slog.Int64("taskID", taskID), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, predictedTask)
}

func (t *Tasks) PredictUncreatedTask(w http.ResponseWriter, r *http.Request) {
	task, err := models.TaskModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	task.CreatedBy = auth.GetUserID(r)

	predictableUserID := groups.GetPredictedUserID(r)
	if predictableUserID != 0 {
		task.CreatedBy = predictableUserID
	}

	predictedTime, err := t.predictionsClient.PredictTask(r.Context(), task)
	if err != nil {
		t.log.Error("error while predicting unreated task", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, &models.PredictedTime{
		PredictedTime: predictedTime,
	})
}

func (t *Tasks) PredictTags(w http.ResponseWriter, r *http.Request) {
	task, err := models.TaskModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, "empty title", http.StatusBadRequest)
		return
	}
	if task.Description == "" {
		http.Error(w, "empty title", http.StatusBadRequest)
		return
	}

	tags, err := t.predictionsClient.PredictTags(r.Context(), task.Title, task.Description)
	if err != nil {
		t.log.Error("error while predicting tags", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, tags)
}

func (t *Tasks) GetTags(w http.ResponseWriter, r *http.Request) {
	tags, err := t.predictionsClient.GetTags(r.Context())
	if err != nil {
		t.log.Error("error while getting tags", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, tags)
}
