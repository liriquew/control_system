package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/liriquew/control_system/internal/models"
	"github.com/liriquew/control_system/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createGroupTask(t *testing.T, ts *suite.Suite, token string, groupID int64, task models.Task) int64 {
	body, _ := json.Marshal(task)
	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks", bytes.NewBuffer(body))
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

func getGroupTask(t *testing.T, ts *suite.Suite, token string, groupID, taskID int64) *models.Task {
	req, _ := http.NewRequest("GET", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(taskID, 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.Task
	json.NewDecoder(resp.Body).Decode(&response)
	return &response
}

func TestCreateGroupTask(t *testing.T) {
	ts := suite.New(t)

	user := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token, group)

	task := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}

	taskID := createGroupTask(t, ts, token, groupID, task)

	retrievedTask := getGroupTask(t, ts, token, groupID, taskID)

	assert.Equal(t, task.Title, retrievedTask.Title)
	assert.Equal(t, task.Description, retrievedTask.Description)
	assert.Equal(t, groupID, retrievedTask.GroupID)
}

func TestGetGroupTaskPermissions(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем первого пользователя
	user1 := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token1 := doSignUpFakeUser(t, ts, user1)

	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token1, group)

	// Создаем задачу для первого пользователя
	task := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	taskID := createGroupTask(t, ts, token1, groupID, task)

	// Успешное получение задачи первым пользователем
	t.Run("Success", func(t *testing.T) {
		retrievedTask := getGroupTask(t, ts, token1, groupID, taskID)
		assert.Equal(t, task.Title, retrievedTask.Title)
		assert.Equal(t, task.Description, retrievedTask.Description)
		assert.Equal(t, groupID, retrievedTask.GroupID)
	})

	// Задача не найдена (несуществующий ID)
	t.Run("Not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"tasks/999999", nil)
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
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(taskID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token2)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Проверяем, что доступ запрещен
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	// Другой участник группы
	t.Run("OK group member", func(t *testing.T) {
		// Регистрируем второго пользователя
		user2 := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		user2, token2 := doSignUpFakeUser(t, ts, user2)

		member := models.GroupMember{
			UserID: user2.ID,
			Role:   "member",
		}
		addGroupMember(t, ts, token1, groupID, member)

		retrievedTask := getGroupTask(t, ts, token2, groupID, taskID)

		assert.Equal(t, task.Title, retrievedTask.Title)
		assert.Equal(t, task.Description, retrievedTask.Description)
		assert.Equal(t, groupID, retrievedTask.GroupID)
	})
}

func TestUpdateGroupTask(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token, group)

	// Создаем задачу для первого пользователя
	task := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task.ID = createGroupTask(t, ts, token, groupID, task)

	// Успешное обновление
	t.Run("Success Admin", func(t *testing.T) {
		updatedTask := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: task.Description,
			PlannedTime: task.PlannedTime,
		}

		body, _ := json.Marshal(updatedTask)
		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(task.ID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		retrievedTask := getGroupTask(t, ts, token, groupID, task.ID)
		assert.Equal(t, updatedTask.Title, retrievedTask.Title)
		assert.Equal(t, updatedTask.Description, retrievedTask.Description)
		assert.Equal(t, updatedTask.PlannedTime, retrievedTask.PlannedTime)
	})

	// Успешное обновление
	t.Run("Success Editor", func(t *testing.T) {
		updatedTask := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: task.Description,
			PlannedTime: task.PlannedTime,
		}

		user2 := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		user2, token2 := doSignUpFakeUser(t, ts, user2)

		member := models.GroupMember{
			UserID: user2.ID,
			Role:   "editor",
		}
		addGroupMember(t, ts, token, groupID, member)

		body, _ := json.Marshal(updatedTask)
		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(task.ID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token2)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		retrievedTask := getGroupTask(t, ts, token, groupID, task.ID)
		assert.Equal(t, updatedTask.Title, retrievedTask.Title)
		assert.Equal(t, updatedTask.Description, retrievedTask.Description)
		assert.Equal(t, updatedTask.PlannedTime, retrievedTask.PlannedTime)
	})

	// Успешное обновление
	t.Run("Forbidden member", func(t *testing.T) {
		updatedTask := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: task.Description,
			PlannedTime: task.PlannedTime,
		}

		user2 := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		user2, token2 := doSignUpFakeUser(t, ts, user2)

		member := models.GroupMember{
			UserID: user2.ID,
			Role:   "member",
		}
		addGroupMember(t, ts, token, groupID, member)

		body, _ := json.Marshal(updatedTask)
		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(task.ID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token2)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Partional Update", func(t *testing.T) {
		updatedTask := models.Task{
			Title: gofakeit.JobTitle(),
		}

		body, _ := json.Marshal(updatedTask)
		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(task.ID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		retrievedTask := getTask(t, ts, token, task.ID)
		assert.Equal(t, updatedTask.Title, retrievedTask.Title)
		assert.Equal(t, task.Description, retrievedTask.Description)
		assert.Equal(t, task.PlannedTime, retrievedTask.PlannedTime)
	})

	t.Run("Bad Update", func(t *testing.T) {
		updatedTask := models.Task{}
		retrievedTask := getTask(t, ts, token, task.ID)

		body, _ := json.Marshal(updatedTask)
		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(task.ID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		msg, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "nothing to update", strings.TrimSpace(string(msg)))

		notUpdatedTask := getTask(t, ts, token, task.ID)

		assert.Equal(t, notUpdatedTask.Title, retrievedTask.Title)
		assert.Equal(t, notUpdatedTask.Description, retrievedTask.Description)
		assert.Equal(t, notUpdatedTask.PlannedTime, retrievedTask.PlannedTime)
	})
}

func TestDeleteGroupTask(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token, group)

	// Успешное удаление
	t.Run("Success Admin", func(t *testing.T) {
		// Создаем задачу для первого пользователя
		task := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: gofakeit.JobDescriptor(),
			PlannedTime: gofakeit.Float64(),
		}
		task.ID = createGroupTask(t, ts, token, groupID, task)

		req, _ := http.NewRequest("DELETE", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(task.ID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверяем, что задача удалена
		req, _ = http.NewRequest("GET", ts.GetURL()+"/api/tasks/"+strconv.FormatInt(task.ID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Успешное удаление
	t.Run("Success Editor", func(t *testing.T) {
		// Создаем задачу для первого пользователя
		task := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: gofakeit.JobDescriptor(),
			PlannedTime: gofakeit.Float64(),
		}
		task.ID = createGroupTask(t, ts, token, groupID, task)

		user2 := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		user2, token2 := doSignUpFakeUser(t, ts, user2)

		member := models.GroupMember{
			UserID: user2.ID,
			Role:   "editor",
		}
		addGroupMember(t, ts, token, groupID, member)

		req, _ := http.NewRequest("DELETE", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(task.ID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token2)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверяем, что задача удалена
		req, _ = http.NewRequest("GET", ts.GetURL()+"/api/tasks/"+strconv.FormatInt(task.ID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Forbidden basic user", func(t *testing.T) {
		// Создаем задачу для первого пользователя
		task := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: gofakeit.JobDescriptor(),
			PlannedTime: gofakeit.Float64(),
		}
		task.ID = createGroupTask(t, ts, token, groupID, task)

		user2 := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		user2, token2 := doSignUpFakeUser(t, ts, user2)

		member := models.GroupMember{
			UserID: user2.ID,
			Role:   "member",
		}
		addGroupMember(t, ts, token, groupID, member)

		req, _ := http.NewRequest("DELETE", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/tasks/"+strconv.FormatInt(task.ID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token2)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestPredictUncreatedTask(t *testing.T) {
	ts := suite.New(t)

	user := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	user2 := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	user2, _ = doSignUpFakeUser(t, ts, user2)

	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token, group)

	member := models.GroupMember{
		UserID: user2.ID,
		Role:   "member",
	}
	body, _ := json.Marshal(member)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/groups/%d/members", ts.GetURL(), groupID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	http.DefaultClient.Do(req)

	start := 10.0
	add := 5.0

	for range 10 {
		task := models.Task{
			Title:       gofakeit.JobTitle(),
			Description: gofakeit.JobDescriptor(),
			AssignedTo:  user2.ID,
			PlannedTime: start,
			ActualTime:  start + mathrand.Float64()*5.0,
			Tags: []int32{
				mathrand.Int31n(100),
				mathrand.Int31n(100),
				mathrand.Int31n(100),
			},
		}
		createGroupTask(t, ts, token, groupID, task)
		start += add
	}

	task := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: ProbablyNormalTaskTime(100),
		Tags:        GetRandTags(4, 100),
	}
	body, _ = json.Marshal(task)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/api/groups/%d/tasks/predict/%d", ts.GetURL(), groupID, user2.ID),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var predictedTime models.PredictedTime
		json.NewDecoder(resp.Body).Decode(&predictedTime)
		assert.NotZero(t, predictedTime.PredictedTime)
	})

	t.Run("Bad request", func(t *testing.T) {
		req, _ := http.NewRequest("GET",
			fmt.Sprintf("%s/api/groups/%d/tasks/predict/%d", ts.GetURL(), groupID, 999999),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("No tasks", func(t *testing.T) {
		user2 := models.User{
			Username: gofakeit.Username(),
			Password: gofakeit.Password(true, true, true, true, false, 10),
		}
		user2, _ = doSignUpFakeUser(t, ts, user2)

		group := models.Group{
			Name:        gofakeit.Company(),
			Description: gofakeit.Sentence(10),
		}
		groupID := createGroup(t, ts, token, group)

		member := models.GroupMember{
			UserID: user2.ID,
			Role:   "member",
		}
		body, _ := json.Marshal(member)

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/groups/%d/members", ts.GetURL(), groupID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		http.DefaultClient.Do(req)

		req, _ = http.NewRequest("GET",
			fmt.Sprintf("%s/api/groups/%d/tasks/predict/%d", ts.GetURL(), groupID, user2.ID),
			bytes.NewBuffer(body),
		)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var predictedTime models.PredictedTime
		json.NewDecoder(resp.Body).Decode(&predictedTime)
		assert.Zero(t, predictedTime.PredictedTime)
	})
}
