package main

import (
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/data"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


func TestPingRoute(t *testing.T) {
	mockAuth := true
	app := setup(mockAuth)
	testRouter := app.Routes()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/ping", nil)
	req.Header.Set("Authorization", "Bearer ")
	testRouter.ServeHTTP(w, req)
	t.Log(w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegisterUser(t *testing.T) {
	testcases := []struct {
		in     []byte
		code   int
		expect string
	}{
		{
			in:     []byte(`{"team":"aces1", "email":"a@b.com", "password":"abc123456"}`),
			code:   http.StatusCreated,
			expect: "user",
		},
		{
			in:     []byte(`{"team":"aces1", "email":"c@b.com", "password":"abc123456"}`),
			code:   http.StatusCreated,
			expect: "user",
		},
	}
	mockAuth := true
	app := setup(mockAuth)

	for _, tcase := range testcases {
		out, code := DoRequest(app, tcase.in, "/v1/users", "", http.MethodPost)
		t.Log(out.String())
		assert.Equal(t, tcase.code, code)
		assert.Equal(t, strings.Contains(out.String(), tcase.expect), true)
	}
	app.Migrations.DoMigrations("down")
}

func TestGetUser(t *testing.T) {

	type model struct {
		team  *data.Team
		users []string
	}
	testcases := []struct {
		user   string
		code   int
		expect string
		model  model
	}{
		{
			user: "a@b",
			model: model{
				team:  &data.Team{Name: "aces"},
				users: []string{"a@b", "b@c"},
			},
			code:   http.StatusCreated,
			expect: "user",
		},
	}

	mockAuth := true
	app := setup(mockAuth)
	for _, tcase := range testcases {

		app.Models.Team.Add(tcase.model.team)
		for _, u := range tcase.model.users {
			userAdd := &data.UserAccount{
				Email: u,
				Team:  &data.Team{Name: tcase.model.team.Name},
			}
			app.Models.UserAccount.Add(userAdd)
		}

		nUser, err := app.Models.UserAccount.GetByEmail(tcase.user)
		assert.Equal(t, err, nil)
		assert.Equal(t, nUser.Email, tcase.user)
	}
	app.Migrations.DoMigrations("down")
}

func TestUpdateUser(t *testing.T) {

	type model struct {
		team  *data.Team
		users []string
	}
	testcases := []struct {
		user        string
		updatedUser string
		code        int
		expect      string
		model       model
	}{
		{
			user:        "a@b",
			updatedUser: "b@v",
			model: model{
				team:  &data.Team{Name: "aces"},
				users: []string{"a@b", "b@c"},
			},
			code:   http.StatusCreated,
			expect: "user",
		},
	}

	mockAuth := true
	app := setup(mockAuth)
	for _, tcase := range testcases {

		app.Models.Team.Add(tcase.model.team)
		for _, u := range tcase.model.users {
			userAdd := &data.UserAccount{
				Email:     u,
				Activated: true,
				Team:      &data.Team{Name: tcase.model.team.Name},
			}
			app.Models.UserAccount.Add(userAdd)
		}

		nUser, err := app.Models.UserAccount.GetByEmail(tcase.user)
		assert.Equal(t, err, nil)
		nUser.Email = tcase.updatedUser
		t.Log(nUser.Version, nUser.Email, nUser.Activated)
		err = app.Models.UserAccount.Update(nUser)
		assert.Equal(t, err, nil)
		assert.Equal(t, nUser.Email, tcase.updatedUser)
		assert.Equal(t, nUser.Version, 2)
	}
	app.Migrations.DoMigrations("down")
}
func TestDeleteUser(t *testing.T) {

	type model struct {
		team  *data.Team
		users []string
	}
	testcases := []struct {
		in 			[]byte
		code        int
		expect      string
		model       model
	}{
		{
			in: []byte(`{"email":"b@c"}`),
			model: model{
				team:  &data.Team{Name: "aces"},
				users: []string{"a@b", "b@c", "c@e"},
			},
			code:   http.StatusOK,
			expect: "user",
		},
	}

	mockAuth := true
	app := setup(mockAuth)
	for _, tcase := range testcases {

		app.Models.Team.Add(tcase.model.team)
		for _, u := range tcase.model.users {
			userAdd := &data.UserAccount{
				Email:     u,
				Activated: true,
				Team:      &data.Team{Name: tcase.model.team.Name},
			}
			app.Models.UserAccount.Add(userAdd)
		}
		out, code := DoRequest(app, tcase.in, "/v1/users", "", http.MethodDelete)
		t.Log(out.String())
		assert.Equal(t, code, tcase.code)
	}
	// app.Migrations.DoMigrations("down")
}

func TestToken(t *testing.T) {

	testcases := []struct {
		user     string
		team     *data.Team
		password string
		code     int
		expect   string
		in       []byte
	}{
		{
			user:     "a@b",
			team:     &data.Team{Name: "aces"},
			password: "abcdef133",
			code:     http.StatusCreated,
			expect:   "user",
			in:       []byte(`{"email":"a@b", "password":"abcdef123"}`),
		},
	}
	mockAuth := true
	app := setup(mockAuth)
	for _, tcase := range testcases {

		app.Models.Team.Add(tcase.team)
		userAdd := &data.UserAccount{
			Email:     tcase.user,
			Activated: true,
			Team:      &data.Team{Name: tcase.team.Name},
		}
		err := app.Models.UserAccount.Add(userAdd)
		assert.Equal(t, err, nil)
		out, code := DoRequest(app, tcase.in, "/v1/tokens/authentication", "", http.MethodPost)
		t.Log(out.String())
		assert.Equal(t, code, tcase.code)
	}
	app.Migrations.DoMigrations("down")
}
