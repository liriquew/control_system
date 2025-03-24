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
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createGraph(t *testing.T, ts *suite.Suite, token string, graph entities.GraphWithNodes) int64 {
	body, err := json.Marshal(graph)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.GetURL()+"/api/groups/"+strconv.FormatInt(graph.GraphInfo.GroupID, 10)+"/graphs", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var createdGraph models.Graph
	err = json.NewDecoder(resp.Body).Decode(&createdGraph)
	require.NoError(t, err)

	return createdGraph.ID
}

func getGraph(t *testing.T, ts *suite.Suite, token string, graphID int64) entities.GraphWithNodes {
	req, err := http.NewRequest("POST", ts.GetURL()+"/api/grahs/"+strconv.FormatInt(graphID, 10), bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var createdGraph entities.GraphWithNodes
	err = json.NewDecoder(resp.Body).Decode(&createdGraph)
	require.NoError(t, err)

	return createdGraph
}

func createNode(t *testing.T, ts *suite.Suite, token string, graphID int64, node models.Node) int64 {
	body, err := json.Marshal(node)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var createdNode models.Node
	err = json.NewDecoder(resp.Body).Decode(&createdNode)
	require.NoError(t, err)

	return createdNode.ID
}

func createDependency(t *testing.T, ts *suite.Suite, token string, graphID, node1ID, node2ID int64) {
	req, err := http.NewRequest("POST", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(node1ID, 10)+"/dependencies/"+strconv.FormatInt(node2ID, 10), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

}

func TestGetGraph(t *testing.T) {
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
	group.ID = createGroup(t, ts, token, group)

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

	// Создаем группу
	graph := entities.GraphWithNodes{
		GraphInfo: models.Graph{
			Name:    gofakeit.BeerName(),
			GroupID: group.ID,
		},
		Nodes: []*models.Node{
			{ID: 1, DependencyNodeIDs: []int64{2}, TaskID: task1.ID},
			{ID: 2, TaskID: task2.ID},
		},
	}
	graphID := createGraph(t, ts, token, graph)

	t.Run("Success", func(t *testing.T) {
		// Получаем граф
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedGraph entities.GraphWithNodes
		err = json.NewDecoder(resp.Body).Decode(&retrievedGraph)
		require.NoError(t, err)

		assert.Equal(t, graph.GraphInfo.Name, retrievedGraph.GraphInfo.Name)

		tasksNodeMap := make(map[int64]int64, 2)
		tasksNodeMap[task1.ID] = 0
		tasksNodeMap[task2.ID] = 0

		var nodeWithDepFound, depFound bool
		var nodeWithDep, depNode *models.Node

		for _, node := range retrievedGraph.Nodes {
			_, ok := tasksNodeMap[node.TaskID]
			assert.True(t, true, ok)
			tasksNodeMap[node.TaskID] = node.ID
			if len(node.DependencyNodeIDs) != 0 {
				assert.False(t, false, nodeWithDepFound)
				nodeWithDep = node
				nodeWithDepFound = true
			} else {
				assert.False(t, false, depFound)
				depNode = node
				depFound = true
			}
		}

		assert.NotNil(t, nodeWithDep)
		assert.NotNil(t, depNode)

		assert.Len(t, nodeWithDep.DependencyNodeIDs, 1)
		assert.Equal(t, nodeWithDep.DependencyNodeIDs[0], depNode.ID)
	})

	t.Run("Forbidden", func(t *testing.T) {
		user := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		_, token := doSignUpFakeUser(t, ts, user)
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestCreateNode(t *testing.T) {
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
	group.ID = createGroup(t, ts, token, group)

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

	// Создаем группу
	graph := entities.GraphWithNodes{
		GraphInfo: models.Graph{
			Name:    gofakeit.BeerName(),
			GroupID: group.ID,
		},
		Nodes: []*models.Node{
			{ID: 1, DependencyNodeIDs: []int64{2}, TaskID: task1.ID},
			{ID: 2, TaskID: task2.ID},
		},
	}
	graph.GraphInfo.ID = createGraph(t, ts, token, graph)

	newTask1 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	newTask1.ID = createTask(t, ts, token, newTask1)

	node := models.Node{
		GraphID:           graph.GraphInfo.ID,
		TaskID:            newTask1.ID,
		DependencyNodeIDs: []int64{},
	}
	t.Run("Success", func(t *testing.T) {
		body, _ := json.Marshal(node)

		req, _ := http.NewRequest("POST", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graph.GraphInfo.ID, 10)+"/nodes", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var createdNode models.Node
		err = json.NewDecoder(resp.Body).Decode(&createdNode)
		require.NoError(t, err)

		assert.NotZero(t, createdNode.ID)
	})

	t.Run("Forbidden", func(t *testing.T) {
		// Регистрируем пользователя
		user := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		_, token := doSignUpFakeUser(t, ts, user)
		body, _ := json.Marshal(node)

		req, _ := http.NewRequest("POST", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graph.GraphInfo.ID, 10)+"/nodes", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestRemoveNode(t *testing.T) {
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
	group.ID = createGroup(t, ts, token, group)

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

	// Создаем группу
	graph := entities.GraphWithNodes{
		GraphInfo: models.Graph{
			Name:    gofakeit.BeerName(),
			GroupID: group.ID,
		},
		Nodes: []*models.Node{
			{ID: 1, DependencyNodeIDs: []int64{2}, TaskID: task1.ID},
			{ID: 2, TaskID: task2.ID},
		},
	}
	graphID := createGraph(t, ts, token, graph)

	task3 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task3.ID = createTask(t, ts, token, task3)

	// Создаем узел
	node := models.Node{
		GraphID:           graphID,
		TaskID:            task3.ID,
		DependencyNodeIDs: []int64{},
	}
	nodeID := createNode(t, ts, token, graphID, node)
	t.Run("Success", func(t *testing.T) {
		// Удаляем узел
		req, _ := http.NewRequest("DELETE", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(nodeID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверяем, что узел удален
		req, _ = http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(nodeID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Forbidden", func(t *testing.T) {
		user := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		_, token := doSignUpFakeUser(t, ts, user)
		// Удаляем узел
		req, _ := http.NewRequest("DELETE", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(nodeID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		// Проверяем, что узел удален
		req, _ = http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(nodeID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestUpdateNode(t *testing.T) {
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
	group.ID = createGroup(t, ts, token, group)

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

	// Создаем группу
	graph := entities.GraphWithNodes{
		GraphInfo: models.Graph{
			Name:    gofakeit.BeerName(),
			GroupID: group.ID,
		},
		Nodes: []*models.Node{
			{ID: 1, DependencyNodeIDs: []int64{2}, TaskID: task1.ID},
			{ID: 2, TaskID: task2.ID},
		},
	}
	graphID := createGraph(t, ts, token, graph)

	task3 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task3.ID = createTask(t, ts, token, task3)
	// Создаем узел
	node := models.Node{
		GraphID:           graphID,
		DependencyNodeIDs: []int64{},
		TaskID:            task3.ID,
	}
	nodeID := createNode(t, ts, token, graphID, node)

	task4 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task4.ID = createTask(t, ts, token, task4)

	// Обновляем узел
	updatedNode := models.Node{
		TaskID: task4.ID,
	}
	t.Run("Success", func(t *testing.T) {
		body, _ := json.Marshal(updatedNode)

		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(nodeID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверяем, что узел обновлен
		req, _ = http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(nodeID, 10), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var retrievedNode models.Node
		err = json.NewDecoder(resp.Body).Decode(&retrievedNode)
		require.NoError(t, err)
	})

	t.Run("Forbidden", func(t *testing.T) {
		// Регистрируем пользователя
		user := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		_, token := doSignUpFakeUser(t, ts, user)
		body, _ := json.Marshal(updatedNode)

		req, _ := http.NewRequest("PATCH", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(nodeID, 10), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestGetDependencies(t *testing.T) {
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

	// Создаем задачи
	task1 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task1ID := createTask(t, ts, token, task1)

	task2 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task2ID := createTask(t, ts, token, task2)

	// Создаем граф
	graph := entities.GraphWithNodes{
		GraphInfo: models.Graph{
			Name:    gofakeit.BeerName(),
			GroupID: groupID,
		},
		Nodes: []*models.Node{
			{
				ID:                1,
				TaskID:            task1ID,
				DependencyNodeIDs: []int64{},
			},
		},
	}
	graphID := createGraph(t, ts, token, graph)

	t.Run("Success Creator", func(t *testing.T) {
		proxyNode := models.Node{
			GraphID: graphID,
			TaskID:  task2ID,
		}
		proxyNode.ID = createNode(t, ts, token, graphID, proxyNode)

		createdNode := models.Node{
			GraphID:           graphID,
			TaskID:            task2ID,
			DependencyNodeIDs: []int64{proxyNode.ID},
		}
		createdNode.ID = createNode(t, ts, token, graphID, createdNode)

		// Получаем зависимости для узла 2
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(createdNode.ID, 10)+"/dependencies", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверяем ответ
		var dependencies models.Node
		err = json.NewDecoder(resp.Body).Decode(&dependencies)
		require.NoError(t, err)

		// Проверяем, что зависимости корректны
		assert.Len(t, dependencies.DependencyNodeIDs, 1)
	})

	t.Run("Forbidden", func(t *testing.T) {
		proxyNode := models.Node{
			GraphID: graphID,
			TaskID:  task2ID,
		}
		proxyNode.ID = createNode(t, ts, token, graphID, proxyNode)

		createdNode := models.Node{
			GraphID:           graphID,
			TaskID:            task2ID,
			DependencyNodeIDs: []int64{proxyNode.ID},
		}
		createdNode.ID = createNode(t, ts, token, graphID, createdNode)
		// Регистрируем пользователя
		user := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		_, token := doSignUpFakeUser(t, ts, user)
		// Получаем зависимости для узла 2
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(createdNode.ID, 10)+"/dependencies", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success Editor", func(t *testing.T) {
		// Регистрируем пользователя
		userEditor := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		_, tokenEditor := doSignUpFakeUser(t, ts, userEditor)

		tokenParsed, err := jwt.Parse(tokenEditor, func(token *jwt.Token) (interface{}, error) {
			return []byte("anyEps"), nil
		})

		require.NoError(t, err)

		claims, ok := tokenParsed.Claims.(jwt.MapClaims)
		require.True(t, ok)

		uidEditor := claims["uid"]

		gm := models.GroupMember{
			UserID: int64(uidEditor.(float64)),
			Role:   "editor",
		}

		body, err := json.Marshal(gm)
		assert.NoError(ts, err)

		req, _ := http.NewRequest("POST", ts.GetURL()+"/api/groups/"+strconv.FormatInt(groupID, 10)+"/members", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		proxyNode := models.Node{
			GraphID: graphID,
			TaskID:  task2ID,
		}
		proxyNode.ID = createNode(t, ts, tokenEditor, graphID, proxyNode)

		createdNode := models.Node{
			GraphID:           graphID,
			TaskID:            task2ID,
			DependencyNodeIDs: []int64{proxyNode.ID},
		}
		createdNode.ID = createNode(t, ts, tokenEditor, graphID, createdNode)

		// Получаем зависимости для узла 2
		req, _ = http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(createdNode.ID, 10)+"/dependencies", nil)
		req.Header.Set("Authorization", "Bearer "+tokenEditor)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверяем ответ
		var dependencies models.Node
		err = json.NewDecoder(resp.Body).Decode(&dependencies)
		require.NoError(t, err)

		// Проверяем, что зависимости корректны
		assert.Len(t, dependencies.DependencyNodeIDs, 1)
	})
}

func TestAddDependency(t *testing.T) {
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
	group.ID = createGroup(t, ts, token, group)

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

	// Создаем группу
	graph := entities.GraphWithNodes{
		GraphInfo: models.Graph{
			Name:    gofakeit.BeerName(),
			GroupID: group.ID,
		},
		Nodes: []*models.Node{
			{ID: 1, DependencyNodeIDs: []int64{2}, TaskID: task1.ID},
			{ID: 2, TaskID: task2.ID},
		},
	}
	graphID := createGraph(t, ts, token, graph)

	task3 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task3.ID = createTask(t, ts, token, task3)

	task4 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task4.ID = createTask(t, ts, token, task4)

	// Создаем два узла
	node1 := models.Node{
		GraphID:           graphID,
		DependencyNodeIDs: []int64{},
		TaskID:            task3.ID,
	}
	node1ID := createNode(t, ts, token, graphID, node1)

	node2 := models.Node{
		GraphID:           graphID,
		DependencyNodeIDs: []int64{},
		TaskID:            task4.ID,
	}
	node2ID := createNode(t, ts, token, graphID, node2)

	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(node1ID, 10)+"/dependencies/"+strconv.FormatInt(node2ID, 10), bytes.NewBuffer([]byte{}))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем, что зависимость добавлена
	req, _ = http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(node1ID, 10)+"/dependencies", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var dependencies models.Node
	err = json.NewDecoder(resp.Body).Decode(&dependencies)
	require.NoError(t, err)

	assert.Len(t, dependencies.DependencyNodeIDs, 1)
	assert.Equal(t, node2ID, dependencies.DependencyNodeIDs[0])
}

func TestRemoveDependency(t *testing.T) {
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
	group.ID = createGroup(t, ts, token, group)

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

	// Создаем группу
	graph := entities.GraphWithNodes{
		GraphInfo: models.Graph{
			Name:    gofakeit.BeerName(),
			GroupID: group.ID,
		},
		Nodes: []*models.Node{
			{ID: 1, DependencyNodeIDs: []int64{2}, TaskID: task1.ID},
			{ID: 2, TaskID: task2.ID},
		},
	}
	graphID := createGraph(t, ts, token, graph)

	task3 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task3.ID = createTask(t, ts, token, task3)
	// Создаем два узла
	node1 := models.Node{
		GraphID:           graphID,
		DependencyNodeIDs: []int64{},
		TaskID:            task3.ID,
	}
	node1ID := createNode(t, ts, token, graphID, node1)

	task4 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
	}
	task4.ID = createTask(t, ts, token, task4)
	node2 := models.Node{
		GraphID:           graphID,
		DependencyNodeIDs: []int64{node1ID},
		TaskID:            task4.ID,
	}
	node2ID := createNode(t, ts, token, graphID, node2)

	// Удаляем зависимость
	req, _ := http.NewRequest("DELETE", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(node2ID, 10)+"/dependencies/"+strconv.FormatInt(node1ID, 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем, что зависимость удалена
	req, _ = http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/nodes/"+strconv.FormatInt(node2ID, 10)+"/dependencies", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var dependencies entities.NodeWithDependency
	err = json.NewDecoder(resp.Body).Decode(&dependencies)
	require.NoError(t, err)

	assert.Len(t, dependencies.DependencyNodeIDs, 0)
}

func TestPredictGraph(t *testing.T) {
	ts := suite.New(t)

	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	_, token := doSignUpFakeUser(t, ts, user)

	user1 := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	user1, _ = doSignUpFakeUser(t, ts, user1)

	user2 := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	user2, _ = doSignUpFakeUser(t, ts, user2)

	group := models.Group{
		Name:        gofakeit.Company(),
		Description: gofakeit.Sentence(10),
	}
	groupID := createGroup(t, ts, token, group)

	task1 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
		UserID:      user1.ID,
	}
	task1ID := createTask(t, ts, token, task1)

	task2 := models.Task{
		Title:       gofakeit.JobTitle(),
		Description: gofakeit.JobDescriptor(),
		PlannedTime: gofakeit.Float64(),
		UserID:      user2.ID,
	}
	task2ID := createTask(t, ts, token, task2)

	graph := entities.GraphWithNodes{
		GraphInfo: models.Graph{
			Name:    gofakeit.BeerName(),
			GroupID: groupID,
		},
		Nodes: []*models.Node{
			{
				ID:                1,
				TaskID:            task1ID,
				DependencyNodeIDs: []int64{},
				AssignedTo:        &user1.ID,
			},
			{
				ID:                2,
				TaskID:            task2ID,
				DependencyNodeIDs: []int64{1},
				AssignedTo:        &user2.ID,
			},
		},
	}
	graphID := createGraph(t, ts, token, graph)

	// Тестируем успешный сценарий
	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/predict", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var predictedGraph entities.PredictedGraph
		err = json.NewDecoder(resp.Body).Decode(&predictedGraph)
		require.NoError(t, err)

		assert.NotNil(t, predictedGraph.Graph)
		assert.NotEmpty(t, predictedGraph.Paths)
	})

	// Тестируем ошибку доступа
	t.Run("Access denied", func(t *testing.T) {
		// Создаем другого пользователя
		otherUser := models.User{
			Username: gofakeit.Username(),
			Password: getSomePassword(),
		}
		_, otherToken := doSignUpFakeUser(t, ts, otherUser)

		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/graphs/"+strconv.FormatInt(graphID, 10)+"/predict", nil)
		req.Header.Set("Authorization", "Bearer "+otherToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	// Тестируем ошибку "граф не найден"
	t.Run("Graph not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.GetURL()+"/api/graphs/999999/predict", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	// Тестируем ошибку "цикл в графе"
	t.Run("Cycle in graph", func(t *testing.T) {
		// Создаем граф с циклом
		cyclicGraph := entities.GraphWithNodes{
			GraphInfo: models.Graph{
				Name:    gofakeit.BeerName(),
				GroupID: groupID,
			},
			Nodes: []*models.Node{
				{
					ID:                1,
					TaskID:            task1ID,
					DependencyNodeIDs: []int64{2},
					AssignedTo:        &user1.ID,
				},
				{
					ID:                2,
					TaskID:            task2ID,
					DependencyNodeIDs: []int64{1},
					AssignedTo:        &user1.ID,
				},
			},
		}
		body, err := json.Marshal(cyclicGraph)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", ts.GetURL()+"/api/groups/"+strconv.FormatInt(graph.GraphInfo.GroupID, 10)+"/graphs", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	})
}
