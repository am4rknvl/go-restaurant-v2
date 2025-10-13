package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"restaurant-system/internal/errors"

	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware that recovers from panics and handles errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				
				appErr := errors.NewInternal("An unexpected error occurred")
				c.JSON(http.StatusInternalServerError, appErr)
				c.Abort()
			}
		}()

		c.Next()

		// Check if there are any errors in the context
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			errors.HandleError(c, err.Err)
		}
	}
}

// RequestLogger logs incoming requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[%s] %s %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Next()
		log.Printf("[%s] %s - Status: %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
	}
}
