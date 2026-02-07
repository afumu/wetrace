package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 是请求失败时的标准化 JSON 响应。
type ErrorResponse struct {
	Success bool     `json:"success"`
	Error   APIError `json:"error"`
}

// APIError 表示返回给客户端的详细错误信息。
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	PayURL  string `json:"pay_url,omitempty"`
}

// SendError 使用给定的 HTTP 状态码和标准化的 JSON 错误载荷进行响应。
func SendError(c *gin.Context, httpStatus int, message string) {
	c.AbortWithStatusJSON(httpStatus, ErrorResponse{
		Success: false,
		Error: APIError{
			Code:    httpStatus,
			Message: message,
		},
	})
}

// SendPaymentRequired 发送一个 402 Payment Required 错误，并附带支付链接。
func SendPaymentRequired(c *gin.Context, message string, payURL string) {
	c.AbortWithStatusJSON(http.StatusPaymentRequired, ErrorResponse{
		Success: false,
		Error: APIError{
			Code:    http.StatusPaymentRequired,
			Message: message,
			PayURL:  payURL,
		},
	})
}

// BadRequest 发送一个 400 Bad Request 错误。
func BadRequest(c *gin.Context, message string) {
	SendError(c, http.StatusBadRequest, message)
}

// NotFound 发送一个 404 Not Found 错误。
func NotFound(c *gin.Context, message string) {
	SendError(c, http.StatusNotFound, message)
}

// InternalServerError 发送一个 500 Internal Server Error 错误。
func InternalServerError(c *gin.Context, message string) {
	SendError(c, http.StatusInternalServerError, message)
}
