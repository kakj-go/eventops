package apistructs

import (
	"fmt"
)

type User struct {
	Name  string
	Email string
	Token string
}

type UserRegisterInfo struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (info UserRegisterInfo) Check() error {
	if info.Name == "" {
		return fmt.Errorf("name can not empty")
	}
	if info.Email == "" {
		return fmt.Errorf("email can not empty")
	}
	if info.Password == "" {
		return fmt.Errorf("password can not empty")
	}
	return nil
}
