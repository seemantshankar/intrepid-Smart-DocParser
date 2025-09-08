package middleware

import (
	"contract-analysis-service/internal/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationMiddleware handles request/response validation
func ValidationMiddleware(validate *validator.Validate) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request
		if c.Request.ContentLength > 0 {
			var body interface{}
			if err := c.ShouldBind(&body); err != nil {
				err := errors.NewAppError("VALIDATION_ERROR", "Invalid request body", err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
			
			if err := validate.Struct(body); err != nil {
				validationErrors := err.(validator.ValidationErrors)
				errs := make([]ValidationError, len(validationErrors))
				for i, ve := range validationErrors {
					errs[i] = ValidationError{
						Field:   ve.Field(),
						Message: ve.Tag(),
					}
				}
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errs})
				return
			}
		}
		
		c.Next()
		
		// Validate response if needed
		if len(c.Errors) > 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": errors.NewAppError("INTERNAL_ERROR", "Failed to process request", c.Errors.Last()),
			})
			return
		}
	}
}
