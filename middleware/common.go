package middleware

import "github.com/gin-gonic/gin"

func SkippedPathPrefixes(c *gin.Context, prefixes ...string) bool {
	if len(prefixes) == 0 {
		return false
	}
	path := c.Request.URL.Path
	pathLen := len(path)
	for _, p := range prefixes {
		// "/" 精确匹配根路径，不作为前缀放行所有路径
		if p == "/" {
			if pathLen == 1 {
				return true
			}
			continue
		}
		if pl := len(p); pathLen >= pl && path[:pl] == p {
			return true
		}
	}
	return false
}

func AllowedPathPrefixes(c *gin.Context, prefixes ...string) bool {
	if len(prefixes) == 0 {
		return true
	}
	path := c.Request.URL.Path
	pathLen := len(path)
	for _, p := range prefixes {
		if pl := len(p); pathLen >= pl && path[:pl] == p {
			return true
		}
	}
	return false
}

func Empty() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
