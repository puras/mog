package web

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const PARAM_ID = "id"

func GetToken(c *gin.Context) string {
	var token string
	auth := c.GetHeader("Authorization")
	prefix := "Bearer "

	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	} else {
		token = auth
	}

	if token == "" {
		token = c.Query("access_token")
	}

	return token
}
