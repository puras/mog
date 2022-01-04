package crud

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

/**
 * @project momo
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-30 12:05
 * @desc
 */
type Page struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

const DEFAULT_PAGE_SIZE = "20"

func GetPageParam(c *gin.Context) (int, int, error) {
	var num = c.DefaultQuery("pageNum", "1")
	var size = c.DefaultQuery("pageSize", DEFAULT_PAGE_SIZE)
	pageNum, err := strconv.Atoi(num)
	if err != nil {
		return 0, 0, err
	}
	pageSize, err := strconv.Atoi(size)
	if err != nil {
		return 0, 0, err
	}
	return pageNum, pageSize, nil
}
