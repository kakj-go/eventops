package actuatorclient

import (
	"gorm.io/gorm"
)

type Client struct {
	client *gorm.DB
}

func NewActuatorClient(client *gorm.DB) *Client {
	return &Client{client: client}
}
