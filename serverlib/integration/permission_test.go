package main

import (
	"net/http"
	"testing"

	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestGetPermissions(t *testing.T) {

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
				Role:  "user",
				Team:  &data.Team{Name: tcase.model.team.Name},
			}
			err := app.Models.UserAccount.Add(userAdd)
			assert.Equal(t, err, nil)
		}
		nUser, err := app.Models.UserAccount.GetByEmail(tcase.user)
		assert.Equal(t, err, nil)
		permissions, err := app.Models.Permission.GetForUser(nUser.ID)
		assert.Equal(t, err, nil)
		assert.Equal(t, len(permissions) > 0, true)
	}
	app.Migrations.DoMigrations("down")
}

func TestAuthorize(t *testing.T) {
	type user struct {
		email string
		role  string
		password string
	}

	type model struct {
		team  *data.Team
		users []user
	}
	testcases := []struct {
		user   string
		code   int
		expect string
		model  model
		in1 []byte
		in2 []byte
	}{
		{
			user: "a@b",
			model: model{
				team:  &data.Team{Name: "aces"},
				users: []user{{"a@b", "anon", "12345abc"}, {"b@c", "user", "12345bcd"}},
			},
			in1: []byte(`{"email":"a@b", "password": "12345abc"}`),
			code:   http.StatusCreated,
			expect: "user",
		},
	}

	mockAuth := false
	app := setup(mockAuth)
	for _, tcase := range testcases {

		app.Models.Team.Add(tcase.model.team)
		for _, u := range tcase.model.users {
			userAdd := &data.UserAccount{
				Email: u.email,
				Role:  u.role,
				Team:  &data.Team{Name: tcase.model.team.Name},
			}
			userAdd.Password.Set(u.password)
			err := app.Models.UserAccount.Add(userAdd)
			assert.Equal(t, err, nil)
		}
		out, code := DoRequest(app, tcase.in1, "/v1/tokens/authentication", "", http.MethodPost)
		token := gjson.Get(out.String(), "authentication_token.plain_text").Str
		t.Log(code)
		out, code = DoRequest(app, []byte(``), "/v1/users", token, http.MethodGet)
		t.Log(out.String())
	}
	app.Migrations.DoMigrations("down")
}
