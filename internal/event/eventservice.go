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

package event

import (
	"encoding/json"
	"eventops/apistructs"
	"eventops/internal/core/client/eventclient"
	"eventops/internal/core/token"
	"eventops/pkg/responsehandler"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) send(c *gin.Context) {
	var eventInfo apistructs.Event
	if err := c.ShouldBind(&eventInfo); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	if err := eventInfo.Check(); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("event check error: %v", err), nil))
		return
	}

	eventInfoContent, err := json.Marshal(eventInfo)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("json Marshal error: %v", err), nil))
		return
	}

	createEvent := eventclient.Event{
		Name:    eventInfo.Name,
		Version: eventInfo.Version,
		Content: string(eventInfoContent),
		Creater: token.GetUserName(c),
		Status:  apistructs.EventCreatedStatus,
	}
	_, err = s.eventDbClient.CreateEvent(nil, &createEvent)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("save event error: %v", err), nil))
		return
	}

	s.process.AddToProcess(createEvent)
	c.JSON(responsehandler.Build(http.StatusOK, "", nil))
}
