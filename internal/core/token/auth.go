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
