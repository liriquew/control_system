package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/liriquew/control_system/api/internal/models"
	"github.com/liriquew/control_system/api/tests/suite"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignUp_Success(t *testing.T) {
	ts := suite.New(t)

	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}

	body, _ := json.Marshal(user)

	req, err := http.NewRequest(
		"POST",
		ts.GetURL()+"/api/signup",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.NotEmpty(t, response.Token)

	tokenParsed, err := jwt.Parse(response.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte("AnyEps"), nil
	})

	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	uid, ok := claims["uid"]

	require.True(t, ok)
	assert.Greater(t, int64(uid.(float64)), int64(0))
}

func TestSignIn_Success(t *testing.T) {
	ts := suite.New(t)

	user := models.User{
		Username: gofakeit.Username(),
		Password: getSomePassword(),
	}
	doSignUp(t, ts, user)

	body, _ := json.Marshal(user)

	req, err := http.NewRequest(
		"POST",
		ts.GetURL()+"/api/signin",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var response struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.NotEmpty(t, response.Token)

	tokenParsed, err := jwt.Parse(response.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte("AnyEps"), nil
	})

	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	uid, ok := claims["uid"]

	require.True(t, ok)
	assert.Greater(t, int64(uid.(float64)), int64(0))
}

func TestSignUp_DuplicateUser(t *testing.T) {
	ts := suite.New(t)

	usrname := gofakeit.Username()

	user := models.User{Username: usrname, Password: "pass"}
	doSignUp(t, ts, user)

	resp := doSignUp(t, ts, user)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestSignIn_InvalidCredentials(t *testing.T) {
	ts := suite.New(t)

	usrname := gofakeit.Username()
	user := models.User{Username: usrname, Password: "correct"}
	doSignUp(t, ts, user)

	invalidUser := models.User{Username: usrname, Password: "wrong"}
	resp := doSignIn(t, ts, invalidUser)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func doSignUp(t *testing.T, ts *suite.Suite, user models.User) *http.Response {
	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	return resp
}

func doSignIn(t *testing.T, ts *suite.Suite, user models.User) *http.Response {
	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", ts.GetURL()+"/api/signin", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	return resp
}

func getSomePassword() string {
	return gofakeit.Password(true, true, true, true, true, 10)
}
