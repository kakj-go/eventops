package userclient

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint
	Name      string
	Email     string
	Password  string
	Salt      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type Client struct {
	client *gorm.DB
}

func NewUserClient(client *gorm.DB) *Client {
	return &Client{client: client}
}

func (client *Client) CreateUser(tx *gorm.DB, user *User) (*User, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (client *Client) GetUserByEmail(tx *gorm.DB, email string) (*User, bool, error) {
	if tx == nil {
		tx = client.client
	}

	var dbUser User
	err := tx.Where("email = ?", email).First(&dbUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &dbUser, true, nil
}

func (client *Client) GetUserByName(tx *gorm.DB, name string) (*User, bool, error) {
	if tx == nil {
		tx = client.client
	}

	var dbUser User
	err := tx.Where("name = ?", name).First(&dbUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &dbUser, true, nil
}
