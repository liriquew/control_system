package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/liriquew/control_system/internal/entities"
	"github.com/liriquew/control_system/internal/models"
	"github.com/liriquew/control_system/tests/suite"

	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createGroup(t *testing.T, ts *suite.Suite, token string, group models.Group) int64 {
	body, err := json.Marshal(group)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.GetURL()+"/api/groups", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var createdGroup models.Group
	err = json.NewDecoder(resp.Body).Decode(&createdGroup)
	require.NoError(t, err)

	assert.NotZero(t, createdGroup.ID)
	assert.Equal(t, group.Name, createdGroup.Name)
	assert.Equal(t, group.Description, createdGroup.Description)

	return createdGroup.ID
}

func TestCreateGroup(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	// Создаем группу
	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}

	body, err := json.Marshal(group)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.GetURL()+"/api/groups", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var createdGroup models.Group
	err = json.NewDecoder(resp.Body).Decode(&createdGroup)
	require.NoError(t, err)

	assert.NotZero(t, createdGroup.ID)
	assert.Equal(t, group.Name, createdGroup.Name)
	assert.Equal(t, group.Description, createdGroup.Description)
}

func TestGetGroup(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	// Создаем группу
	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token, group)

	t.Run("Success", func(t *testing.T) {
		// Получаем группу
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedGroup models.Group
		err = json.NewDecoder(resp.Body).Decode(&retrievedGroup)
		require.NoError(t, err)

		assert.Equal(t, group.Name, retrievedGroup.Name)
		assert.Equal(t, group.Description, retrievedGroup.Description)
	})

	t.Run("Forbidden", func(t *testing.T) {
		user := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		_, token := doSignUpFakeUser(t, ts, user)
		// Получаем группу
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestListUserGroups(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	// Создаем несколько групп
	group1 := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	group2 := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	createGroup(t, ts, token, group1)
	createGroup(t, ts, token, group2)

	// Получаем список групп
	req, _ := http.NewRequest("GET", ts.GetURL()+"/api/groups", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var groups []models.Group
	err = json.NewDecoder(resp.Body).Decode(&groups)
	require.NoError(t, err)

	assert.Len(t, groups, 2)
	assert.Equal(t, group1.Name, groups[0].Name)
	assert.Equal(t, group2.Name, groups[1].Name)
}

func TestDeleteGroup(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	// Создаем группу
	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token, group)

	t.Run("Success", func(t *testing.T) {
		// Удаляем группу
		req, _ := http.NewRequest("DELETE", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверяем, что группа удалена
		req, _ = http.NewRequest("GET", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Forbidden", func(t *testing.T) {
		// Регистрируем пользователя
		user := models.User{
			Username: gofakeit.Username(),
			Password: gofakeit.Password(true, true, true, true, false, 10),
		}
		_, token := doSignUpFakeUser(t, ts, user)

		// Удаляем группу
		req, _ := http.NewRequest("DELETE", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestUpdateGroup(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	// Создаем группу
	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token, group)

	// Обновляем группу
	updatedGroup := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	body, _ := json.Marshal(updatedGroup)

	req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем, что группа обновлена
	req, _ = http.NewRequest("GET", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var retrievedGroup models.Group
	err = json.NewDecoder(resp.Body).Decode(&retrievedGroup)
	require.NoError(t, err)

	assert.Equal(t, updatedGroup.Name, retrievedGroup.Name)
	assert.Equal(t, updatedGroup.Description, retrievedGroup.Description)
}

func TestAddAndRemoveGroupMember(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем двух пользователей
	user1 := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	_, token1 := doSignUpFakeUser(t, ts, user1)

	user2 := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	user2, _ = doSignUpFakeUser(t, ts, user2)

	// Создаем группу от имени первого пользователя
	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token1, group)

	// Добавляем второго пользователя в группу
	member := models.GroupMember{
		UserID: user2.ID,
		Role:   "member",
	}
	body, _ := json.Marshal(member)

	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/members", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token1)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Удаляем второго пользователя из группы
	req, _ = http.NewRequest("DELETE", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/members/"+strconv.FormatInt(user2.ID, 10), nil)
	req.Header.Set("Authorization", "Bearer "+token1)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestListGroupMembers(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем двух пользователей
	user1 := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	_, token1 := doSignUpFakeUser(t, ts, user1)

	user2 := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	user2ID, _ := doSignUpFakeUser(t, ts, user2)

	// Создаем группу от имени первого пользователя
	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token1, group)

	// Добавляем второго пользователя в группу
	member := models.GroupMember{
		UserID: user2ID.ID,
		Role:   "member",
	}
	body, _ := json.Marshal(member)

	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/members", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token1)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Получаем список участников группы
	req, _ = http.NewRequest("GET", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/members", nil)
	req.Header.Set("Authorization", "Bearer "+token1)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var members []models.User
	err = json.NewDecoder(resp.Body).Decode(&members)
	require.NoError(t, err)

	assert.Len(t, members, 2)
	assert.Equal(t, user1.Username, members[0].Username)
	assert.Equal(t, user2.Username, members[1].Username)
}

func TestChangeMemberRole(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем двух пользователей
	user1 := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	_, token1 := doSignUpFakeUser(t, ts, user1)

	user2 := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	user2ID, _ := doSignUpFakeUser(t, ts, user2)

	// Создаем группу от имени первого пользователя
	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token1, group)

	// Добавляем второго пользователя в группу
	member := models.GroupMember{
		UserID: user2ID.ID,
		Role:   "member",
	}
	body, _ := json.Marshal(member)

	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/members", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token1)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Меняем роль второго пользователя
	newRole := "admin"
	req, _ = http.NewRequest("PATCH", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/members/"+strconv.FormatInt(user2ID.ID, 10)+"/role?newRole="+newRole, nil)
	req.Header.Set("Authorization", "Bearer "+token1)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateAndListGroupGraphs(t *testing.T) {
	ts := suite.New(t)

	// Регистрируем пользователя
	user := models.User{
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	// Создаем группу
	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token, group)

	task1 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}

	task1.ID = createTask(t, ts, token, task1)

	task2 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}

	task2.ID = createTask(t, ts, token, task2)

	// Создаем граф
	graph := entities.GraphWithNodes{
		GraphInfo: models.Graph{
			Name: gofakeit.BeerName(),
		},
		Nodes: []*models.Node{
			{ID: 1, DependencyNodeIDs: []int64{2}, TaskID: task1.ID},
			{ID: 2, TaskID: task1.ID},
		},
	}
	body, _ := json.Marshal(graph)

	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/graphs", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var createdGraph models.Graph
	err = json.NewDecoder(resp.Body).Decode(&createdGraph)
	require.NoError(t, err)

	assert.NotZero(t, createdGraph.ID)

	// Получаем список графов
	req, _ = http.NewRequest("GET", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/graphs", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var graphs []entities.GraphWithNodes
	err = json.NewDecoder(resp.Body).Decode(&graphs)

	require.NoError(t, err)

	assert.Len(t, graphs, 1)
}
