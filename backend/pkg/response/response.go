package response

import "github.com/gin-gonic/gin"

// Envelope adalah format response JSON konsisten untuk seluruh API.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func OK(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Envelope{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data interface{}) {
	OK(c, 201, message, data)
}

// Fail mengirim response error dan menghentikan chain handler.
func Fail(c *gin.Context, status int, message string, err interface{}) {
	c.AbortWithStatusJSON(status, Envelope{Success: false, Message: message, Error: err})
}
