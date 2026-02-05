package response

import (
	"github.com/gin-gonic/gin"
)

type responseSuccess struct {
	Success bool `json:"success"`
	Code    int  `json:"code"`
	Data    any  `json:"data"`
}

type responseError struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Error   string `json:"error"`
}

func ResponseSuccess(c *gin.Context, code int, data any) {
	c.JSON(code, responseSuccess{
		Success: true,
		Code:    code,
		Data:    data,
	})
}

func ResponseError(c *gin.Context, code int, err error) {
	c.JSON(code, responseError{
		Success: false,
		Code:    code,
		Error:   err.Error(),
	})
}
