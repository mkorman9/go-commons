package web

import "github.com/gin-gonic/gin"

func BindJSONBody(c *gin.Context, val interface{}) (bool, []Cause) {
	if err := c.ShouldBindJSON(val); err != nil {
		return false, []Cause{FieldError("body", "json")}
	}

	return ValidateStruct(val)
}
