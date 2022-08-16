package pagehelper

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetPageAndPageSizeFromGinContext(c *gin.Context) (int64, int64, error) {
	pageString := c.Query("page")
	PageSizeString := c.Query("pageSize")

	page, err := strconv.ParseInt(pageString, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	pageSize, err := strconv.ParseInt(PageSizeString, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return page, pageSize, nil
}
