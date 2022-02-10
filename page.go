package mog

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

const DEFAULT_PAGE = 1
const DEFAULT_PAGE_SIZE = 20

func GetPageParam(c *gin.Context) (int, int) {
	var num = c.DefaultQuery("pageNum", strconv.Itoa(DEFAULT_PAGE))
	var size = c.DefaultQuery("pageSize", strconv.Itoa(DEFAULT_PAGE_SIZE))
	pageNum, err := strconv.Atoi(num)
	if err != nil {
		// return 0, 0, err
		pageNum = DEFAULT_PAGE
	}
	pageSize, err := strconv.Atoi(size)
	if err != nil {
		// return 0, 0, err
		pageSize = DEFAULT_PAGE_SIZE
	}
	return pageNum, pageSize
}

func GetPageOffset(num, size int) int {
	if num < 1 {
		num = 1
	}
	return (num - 1) * size
}
