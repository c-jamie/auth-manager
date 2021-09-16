package api

import (
	"net/http"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/data"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/validator"
)

func (app *Application) registerUserHandeler(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Team     string `json:"team"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		app.badRequest(c, err)
		return
	}

	team := &data.Team{Name: input.Team}
	user := &data.UserAccount{
		Email:     input.Email,
		Activated: true,
		Team:      team,
	}

	err := user.Password.Set(input.Password)
	if err != nil {
		app.badRequest(c, err)
		return
	}

	v := validator.New()

	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}

	data.ValidateUser(v, user)

	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}

	err = app.Models.UserAccount.Add(user)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "duplicate"):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(c, v.Errors)
			return
		default:
			app.badRequest(c, err)
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})

}

func (app *Application) deleteUserHandeler(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		app.badRequest(c, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}

	user, err := app.Models.UserAccount.GetByEmail(input.Email)

	if err != nil {
		app.badRequest(c, err)
		return
	}

	err = app.Models.UserAccount.Delete(user)

	if err != nil {
		app.badRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (app *Application) getUserHandeler(c *gin.Context) {

	authorizationHeader := c.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		c.JSON(http.StatusOK, gin.H{"user": data.AnonUser})
		return
	}

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		app.invalidAuthenticationTokenResponse(c)
		return
	}

	token := headerParts[1]
	if token == "" {
		app.invalidAuthenticationTokenResponse(c)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, token); !v.Valid() {
		app.invalidAuthenticationTokenResponse(c)
		return
	}

	users, err := app.Models.UserAccount.GetForToken(data.ScopeLogin, token)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "no records"):
			app.invalidAuthenticationTokenResponse(c)
			return
		default:
			app.badRequest(c, err)
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"user": users})
}

func (app *Application) createAuthenticationTokenHandeler(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		app.badRequest(c, err)
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}

	user, err := app.Models.UserAccount.GetByEmail(input.Email)
	if err != nil {
		app.invalidCredentialsResponse(c)
		return
	}

	token, err := app.Models.Token.New(user.ID, 24*time.Hour, data.ScopeLogin)
	if err != nil {
		app.badRequest(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"authentication_token": token})
}