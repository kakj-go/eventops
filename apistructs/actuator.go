package apistructs

import (
	"time"
)

type Actuator struct {
	Name    string   `json:"name"`
	Creater string   `json:"creater"`
	Type    string   `json:"type"`
	Status  string   `json:"status"`
	Tags    []string `json:"tags"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
