package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (app *Application) failedValidationResponse(c *gin.Context, errors map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errors})
}

func (app *Application) badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
}

func (app *Application) invalidCredentialsResponse(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"errors": "invalid credentials"})
}

func (app *Application) invalidAuthenticationTokenResponse(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"errors": "invalid or missing authentication token"})
}
