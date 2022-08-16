package client

import (
	"gorm.io/gorm"
)

type Client struct {
	client *gorm.DB
}

func NewFlowClient(client *gorm.DB) *Client {
	return &Client{client: client}
}
