/*
 * Copyright 2022 The kakj-go Authors.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package token

import (
	"eventops/conf"
	"eventops/pkg/responsehandler"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

const AuthHeader = "Authorization"
const UserNameHeader = "event_ops_username"
const Bearer = "Bearer"

func isWriteUrlList(url string) bool {
	for _, writeUrl := range conf.GetUc().Auth.WhiteUrlList {
		if strings.EqualFold(url, writeUrl) {
			return true
		}
	}
	return false
}

func BuildTokenHeaderValue(token string) string {
	return fmt.Sprintf("%s %s", Bearer, token)
}

func LoginAuth(c *gin.Context) {
	if isWriteUrlList(c.Request.URL.Path) {
		c.Next()
		return
	}

	tokenValue := strings.Split(c.GetHeader(AuthHeader), Bearer+" ")
	if len(tokenValue) != 2 {
		c.JSON(responsehandler.Build(http.StatusUnauthorized, fmt.Sprintf("auth fail: not get token in header %v", AuthHeader), nil))
		c.Abort()
		return
	}

	jwtToken, name, err := ParseLoginToken(tokenValue[1], conf.GetLoginTokenSignature())
	if err != nil {
		log.Errorf("parse login token error %v", err)
		c.JSON(responsehandler.Build(http.StatusUnauthorized, fmt.Sprintf("auth fail: %v", err), nil))
		c.Abort()
		return
	}

	c.Request.Header.Add(UserNameHeader, name)
	defer c.Request.Header.Del(UserNameHeader)

	c.Next()

	newToken, err := UpdateExpiresTime(jwtToken, time.Now().Add(conf.GetLoginTokenExpiresTime()), conf.GetLoginTokenSignature())
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("update token expires time error %v", err), nil))
		c.Abort()
		return
	}
	SetToken(c, newToken)
}
