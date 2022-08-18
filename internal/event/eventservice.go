package event

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"tiggerops/apistructs"
	"tiggerops/internal/event/client/eventclient"
	"tiggerops/pkg/responsehandler"
	"tiggerops/pkg/token"
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
