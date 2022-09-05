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

package userclient

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id        uint
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
