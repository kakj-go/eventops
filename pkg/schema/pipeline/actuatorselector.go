package pipeline

import (
	"fmt"
	"strings"
)

type ActuatorSelector struct {
	Tags []string `yaml:"tags,omitempty"`
}

func (a ActuatorSelector) check() error {
	if a.Tags == nil {
		return nil
	}

	for _, tag := range a.Tags {
		if strings.TrimSpace(tag) == "" {
			return fmt.Errorf("actuatorSelector tag can not empty")
		}
	}

	return nil
}
