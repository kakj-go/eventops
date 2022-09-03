/*
 * Copyright (c) 2021 Terminus, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
