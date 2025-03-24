package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/liriquew/control_system/internal/models"
	"github.com/liriquew/control_system/tests/suite"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTask(t *testing.T) {
	ts := suite.New(t)

	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	t.Run("Success", func(t *testing.T) {
		task := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: gofakeit.JobDescriptor(),
			PlannedTime: gofakeit.Float64(),
		}

		taskID := createTask(t, ts, token, task)
		assert.NotZero(t, taskID)
	})

	t.Run("Invalid Title", func(t *testing.T) {
		task := models.Task{
			Title:       "",
			Description: gofakeit.JobDescriptor(),
			PlannedTime: gofakeit.Float64(),
		}

		body, _ := json.Marshal(task)
		req, _ := http.NewRequest("POST", ts.GetURL()+"/api/tasks", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		msg, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "invalid title", strings.TrimSpace(string(msg)))
	})

	t.Run("Invalid Description", func(t *testing.T) {
		task := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: "",
			PlannedTime: gofakeit.Float64(),
		}

		body, _ := json.Marshal(task)
		req, _ := http.NewRequest("POST", ts.GetURL()+"/api/tasks", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		msg, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "invalid description", strings.TrimSpace(string(msg)))
	})
}

func TestGetTask(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем первого пользователя
	user1 := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token1 := doSignUpFakeUser(t, ts, user1)

	// Создаем задачу для первого пользователя
	task := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	taskID := createTask(t, ts, token1, task)

	// Успешное получение задачи первым пользователем
	t.Run("Success", func(t *testing.T) {
		retrievedTask := getTask(t, ts, token1, taskID)
		assert.Equal(t, task.Title, retrievedTask.Title)
		assert.Equal(t, task.Description, retrievedTask.Description)
	})

	// Задача не найдена (несуществующий ID)
	t.Run("Not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/tasks/999999", nil)
		req.Header.Set("Authorization", "Bearer "+token1)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Попытка получить чужую задачу
	t.Run("Access denied for another user", func(t *testing.T) {
		// Регистрируем второго пользователя
		user2 := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		_, token2 := doSignUpFakeUser(t, ts, user2)

		// Второй пользователь пытается получить задачу первого пользователя
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/tasks/"+strconv.FormatInt(taskID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token2)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Проверяем, что доступ запрещен
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestGetTaskList(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	// Создаем несколько задач
	for i := 0; i < 5; i++ {
		task := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: gofakeit.JobDescriptor(),
			PlannedTime: gofakeit.Float64(),
		}
		createTask(t, ts, token, task)
	}

	// Успешное получение списка
	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/tasks?limit=3&offset=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tasks []models.Task

		json.NewDecoder(resp.Body).Decode(&tasks)
		assert.Len(t, tasks, 3)
	})
}

func TestUpdateTask(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	// Создаем задачу
	task := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	taskID := createTask(t, ts, token, task)

	// Успешное обновление
	t.Run("Success", func(t *testing.T) {
		updatedTask := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: task.Description,
			PlannedTime: task.PlannedTime,
		}

		body, _ := json.Marshal(updatedTask)
		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/tasks/"+strconv.FormatInt(taskID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		retrievedTask := getTask(t, ts, token, taskID)
		assert.Equal(t, updatedTask.Title, retrievedTask.Title)
		assert.Equal(t, updatedTask.Description, retrievedTask.Description)
		assert.Equal(t, updatedTask.PlannedTime, retrievedTask.PlannedTime)
	})

	t.Run("Partional Update", func(t *testing.T) {
		updatedTask := models.Task{
			Title: gofakeit.JobTitle(),
		}

		body, _ := json.Marshal(updatedTask)
		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/tasks/"+strconv.FormatInt(taskID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		retrievedTask := getTask(t, ts, token, taskID)
		assert.Equal(t, updatedTask.Title, retrievedTask.Title)
		assert.Equal(t, task.Description, retrievedTask.Description)
		assert.Equal(t, task.PlannedTime, retrievedTask.PlannedTime)
	})

	t.Run("Bad Update", func(t *testing.T) {
		updatedTask := models.Task{}
		retrievedTask := getTask(t, ts, token, taskID)

		body, _ := json.Marshal(updatedTask)
		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/tasks/"+strconv.FormatInt(taskID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		msg, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "nothing to update (empty fields)", strings.TrimSpace(string(msg)))

		notUpdatedTask := getTask(t, ts, token, taskID)

		assert.Equal(t, notUpdatedTask.Title, retrievedTask.Title)
		assert.Equal(t, notUpdatedTask.Description, retrievedTask.Description)
		assert.Equal(t, notUpdatedTask.PlannedTime, retrievedTask.PlannedTime)
	})
}

func TestDeleteTask(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	// Создаем задачу
	task := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	taskID := createTask(t, ts, token, task)

	// Успешное удаление
	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", ts.GetURL()+"/api/tasks/"+strconv.FormatInt(taskID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверяем, что задача удалена
		req, _ = http.NewRequest("GET", ts.GetURL()+"/api/tasks/"+strconv.FormatInt(taskID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func createTask(t *testing.T, ts *suite.Suite, token string, task models.Task) int64 {
	body, _ := json.Marshal(task)
	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		ID json.Number `json:"id"`
	}

	json.NewDecoder(resp.Body).Decode(&response)
	id, err := response.ID.Int64()
	require.NoError(t, err)
	return id
}

func getTask(t *testing.T, ts *suite.Suite, token string, taskID int64) *models.Task {
	req, _ := http.NewRequest("GET", ts.GetURL()+"/api/tasks/"+strconv.FormatInt(taskID, 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var task models.Task
	json.NewDecoder(resp.Body).Decode(&task)
	return &task
}

func doSignUpFakeUser(t *testing.T, ts *suite.Suite, user models.User) (models.User, string) {
	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)

	var response struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&response)

	tokenParsed, _ := jwt.Parse(response.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte("anyEps"), nil
	})

	claims, _ := tokenParsed.Claims.(jwt.MapClaims)

	uid := claims["uid"]
	user.ID = int64(uid.(float64))
	return user, response.Token
}
