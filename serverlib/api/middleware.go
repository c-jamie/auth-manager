package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/data"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/permission"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/validator"
	"github.com/gin-gonic/gin"
)

type contextKey string
const userContextKey = contextKey("user")

// Middleware is the interface used to control permissioning for the app
type Middleware interface {
	// Authenticate determines if a user is allowed visibility on an object
	Authenticate(c *gin.Context)
	// Authorize determines if current subject has been authorized to read an object
	Authorize(code string) gin.HandlerFunc
}

type middleware struct {
	SQLMngrCentral string
	DB             *sql.DB
	Permissions    *data.PermissionModel
}

func NewMiddleware(central string, db *sql.DB) *middleware {
	pmanager := permission.PermissionManager{}
	permissions := data.PermissionModel{Manager: pmanager, DB: db}
	return &middleware{SQLMngrCentral: central, DB: db, Permissions: &permissions}
}

func (mi *middleware) Authenticate(c *gin.Context) {
	authorizationHeader := c.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		mi.contextSetUser(c, data.AnonUser)
		c.Next()
		return
	}

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		mi.invalidAuthenticationTokenResponse(c)
		c.Abort()
		return
	}

	token := headerParts[1]
	if token == "" {
		mi.invalidAuthenticationTokenResponse(c)
		c.Abort()
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, token); !v.Valid() {
		mi.invalidAuthenticationTokenResponse(c)
		c.Abort()
		return
	}

	userMod := data.UserAccountModel{DB: mi.DB}
	user, err := userMod.GetForToken(data.ScopeLogin, token)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "no records"):
			mi.invalidAuthenticationTokenResponse(c)
			c.Abort()
			return
		default:
			mi.badRequest(c, err)
			c.Abort()
			return
		}
	}
	fmt.Println("=== here ===")
	mi.contextSetUser(c, user)
	c.Next()
}

// Authorize determines if current subject has been authorized to take an action on an object.
func (mi *middleware) Authorize(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := mi.contextGetUser(c)
		perm, err := mi.Permissions.GetForUser(user.ID)
		if err != nil {
			mi.badRequest(c, err)
			c.Abort()
			return
		}
		fmt.Println(perm, perm.Include(code))
		if !perm.Include(code) {
			mi.notPermittedResponse(c, code)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (app *middleware) invalidAuthenticationTokenReponse(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
}

func (app *middleware) notPermittedResponse(c *gin.Context, code string) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "code " + code + " not permitted"})
}

func (app *middleware) badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func (app *middleware) contextSetUser(r *gin.Context, user *data.UserAccount) *gin.Context {
	r.Set(string(userContextKey), user)
	return r
}

func (app *middleware) invalidAuthenticationTokenResponse(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"errors": "invalid or missing authentication token"})
}

func (app *middleware) contextGetUser(r *gin.Context) *data.UserAccount {
	user, ok := r.Value(string(userContextKey)).(*data.UserAccount)
	if !ok {
		panic("missing user value in request")
	}
	return user
}