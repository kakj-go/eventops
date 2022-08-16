package uc

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"tiggerops/conf"
	"tiggerops/internal/uc/client/userclient"
	"tiggerops/pkg/responsehandler"
	"tiggerops/pkg/token"
	"time"
)

type User struct {
	Name  string
	Email string
	Token string
}

func (u *Service) me(c *gin.Context) {
	name := token.GetUserName(c)

	dbUser, find, err := u.userDbClient.GetUserByName(nil, name)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}
	if !find {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, "not find user", nil))
		return
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", User{
		Name:  dbUser.Name,
		Email: dbUser.Email,
	}))
}

type UserRegisterInfo struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Check todo check name email and password
func (info UserRegisterInfo) Check() error {
	return nil
}

func (u *Service) register(c *gin.Context) {
	var userRegisterInfo UserRegisterInfo
	if err := c.ShouldBind(&userRegisterInfo); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	if err := userRegisterInfo.Check(); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	dbUser, find, err := u.userDbClient.GetUserByName(nil, userRegisterInfo.Name)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}
	if !find {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("name %v is already in use", userRegisterInfo.Name), nil))
		return
	}

	dbUser, find, err = u.userDbClient.GetUserByName(nil, userRegisterInfo.Email)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}
	if !find {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("email %v is already in use", userRegisterInfo.Name), nil))
		return
	}

	var salt = token.Salt()
	var user = userclient.User{
		Name:     userRegisterInfo.Name,
		Email:    userRegisterInfo.Email,
		Password: encryptPassword(userRegisterInfo.Password, salt),
		Salt:     salt,
	}
	dbUser, err = u.userDbClient.CreateUser(nil, &user)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	c.JSON(responsehandler.Build(http.StatusOK, "register success", User{
		Name:  dbUser.Name,
		Email: dbUser.Email,
	}))
}

func encryptPassword(password string, salt string) string {
	md5Inst := md5.New()
	md5Inst.Write([]byte(password))
	md5Inst.Write([]byte(salt))
	return hex.EncodeToString(md5Inst.Sum(nil))
}

type LoginInfo struct {
	NameOrEmail string `json:"username"`
	Password    string `json:"password"`
}

func (u *Service) login(c *gin.Context) {
	var loginInfo LoginInfo
	if err := c.ShouldBind(&loginInfo); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	dbUser, find, err := u.userDbClient.GetUserByName(nil, loginInfo.NameOrEmail)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}
	if !find {
		dbUser, find, err = u.userDbClient.GetUserByEmail(nil, loginInfo.NameOrEmail)
		if err != nil {
			c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
			return
		}
	}

	if !find {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("username or password error"), nil))
		return
	}

	if dbUser.Password != encryptPassword(loginInfo.Password, dbUser.Salt) {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("username or password error"), nil))
		return
	}

	loginToken, err := token.GenLoginToken(dbUser.Name, conf.GetLoginTokenSignature(), time.Now().Add(conf.GetLoginTokenExpiresTime()))
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("create token error: %v", err), nil))
		return
	}
	token.SetToken(c, loginToken)

	c.JSON(responsehandler.Build(http.StatusOK, "login success", User{
		Name:  dbUser.Name,
		Email: dbUser.Email,
		Token: loginToken,
	}))
}
