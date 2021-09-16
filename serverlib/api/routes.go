package api

import (
	"io"
	"time"

	"github.com/c-jamie/sql-manager-acc-auth/serverlib/log"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logrus(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now().UTC()
		path := c.Request.URL.Path
		c.Next()
		end := time.Now().UTC()
		latency := end.Sub(start)
		logger.WithFields(logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"duration":   latency,
			"user_agent": c.Request.UserAgent(),
		}).Info()
	}
}

// WriteFunc convert func to io.Writer.
type writeFunc func([]byte) (int, error)

func (fn writeFunc) Write(data []byte) (int, error) {
	return fn(data)
}

func newLogrusWrite() io.Writer {
	return writeFunc(func(data []byte) (int, error) {
		log.Debugf("%s", data)
		return 0, nil
	})
}

// Routes returns all routes for the application
func (app *Application) Routes() *gin.Engine {

	gin.SetMode(app.Config.GinMode)
	gin.DefaultWriter = newLogrusWrite()
	router := gin.New()
	router.Use(Logrus(log.Log))

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found", "uri": c.Request.RequestURI})
	})

	public := router.Group("/" + app.Config.Version)

	public.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": app.Config.Version, 
			"build_version": app.Config.BuildVersion,
			"api_version": app.Config.APIVerion,
		})
	})

	private := router.Group("/" + app.Config.Version)
	authenticate := func() gin.HandlerFunc { return app.Middleware.Authenticate }

	private.Use(authenticate())
	private.DELETE("/users", app.Middleware.Authorize("/users-write"), app.deleteUserHandeler)
	private.POST("/users", app.Middleware.Authorize("/users-write"), app.registerUserHandeler)
	private.GET("/users", app.Middleware.Authorize("/users-read"), app.getUserHandeler)
	private.POST("/tokens/authentication", app.createAuthenticationTokenHandeler)

	return router
}
