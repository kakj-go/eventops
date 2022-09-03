package register

import (
	"eventops/apistructs"
	"eventops/internal/core/client/actuatorclient"
	"eventops/internal/core/token"
	"eventops/pkg/responsehandler"
	"eventops/pkg/schema/actuator"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"net/http"
)

type ApplyActuatorRequest struct {
	ActuatorContent string `json:"actuatorContent"`
}

func (r *Service) ApplyActuator(c *gin.Context) {
	var applyInfo ApplyActuatorRequest
	if err := c.ShouldBind(&applyInfo); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	var actuatorInfo = &actuator.Client{}
	err := yaml.Unmarshal([]byte(applyInfo.ActuatorContent), actuatorInfo)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("yaml content unmarshal error: %v", err), nil))
		return
	}

	if err := actuatorInfo.Mutating(); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	if err := actuatorInfo.Check(); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	yamlContent, err := yaml.Marshal(actuatorInfo)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("pipeline yaml content marshal error: %v", err), nil))
		return
	}
	applyContent := string(yamlContent)

	dbActuator, find, err := r.actuatorClient.GetActuator(nil, actuatorInfo.Name, token.GetUserName(c))
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("get actuator error: %v", err), nil))
		return
	}

	err = r.dbClient.Transaction(func(tx *gorm.DB) error {
		var dbActuatorId uint64
		if !find {
			var create = actuatorclient.Actuator{
				Name:        actuatorInfo.Name,
				Creater:     token.GetUserName(c),
				Type:        actuatorInfo.GetType(),
				Content:     applyContent,
				ClientId:    actuatorInfo.GetTunnelClientID(),
				ClientToken: actuatorInfo.GetTunnelClientToken(),
			}
			if _, err := r.actuatorClient.CreateActuator(tx, &create); err != nil {
				return err
			}
			dbActuatorId = create.Id
		} else {
			dbActuator.Content = applyContent
			dbActuator.Type = actuatorInfo.GetType()

			dbActuator.ClientId = actuatorInfo.GetTunnelClientID()
			dbActuator.ClientToken = actuatorInfo.GetTunnelClientToken()
			if _, err := r.actuatorClient.UpdateActuator(tx, dbActuator); err != nil {
				return err
			}
			dbActuatorId = dbActuator.Id
		}

		if err := r.actuatorClient.DeleteActuatorTags(tx, actuatorInfo.Name, token.GetUserName(c)); err != nil {
			return err
		}
		var actuatorTags []actuatorclient.ActuatorTag
		for _, tag := range actuatorInfo.Tags {
			actuatorTags = append(actuatorTags, actuatorclient.ActuatorTag{
				ActuatorName:    actuatorInfo.Name,
				ActuatorCreater: token.GetUserName(c),
				ActuatorType:    actuatorInfo.GetType(),
				Tag:             tag,
				ActuatorId:      dbActuatorId,
			})
		}
		err := r.actuatorClient.BatchCreateActuatorTags(tx, actuatorTags)
		if err != nil {
			return err
		}

		r.dialerServer.AddAuthInfo(actuatorInfo.GetTunnelClientID(), token.GetUserName(c), actuatorInfo.GetTunnelClientToken())
		return nil
	})
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("save actuator error: %v", err), nil))
		return
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", nil))
}

type DeleteActuatorUrlQuery struct {
	Name string `uri:"name"`
}

func (r *Service) DeleteActuator(c *gin.Context) {
	var deleteQuery = DeleteActuatorUrlQuery{}
	if err := c.ShouldBindUri(&deleteQuery); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get name from uri error: %v", err), nil))
		return
	}

	err := r.dbClient.Transaction(func(tx *gorm.DB) error {
		dbActuator, find, err := r.actuatorClient.GetActuator(nil, deleteQuery.Name, token.GetUserName(c))
		if err != nil {
			return err
		}
		if !find {
			return nil
		}

		err = r.actuatorClient.DeleteActuator(tx, dbActuator.Name, token.GetUserName(c))
		if err != nil {
			return err
		}

		err = r.actuatorClient.DeleteActuatorTags(tx, dbActuator.Name, token.GetUserName(c))
		if err != nil {
			return err
		}

		r.dialerServer.DeleteAuthInfo(dbActuator.ClientId, token.GetUserName(c))
		return nil
	})
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("delete actuator error: %v", err), nil))
		return
	}
	c.JSON(responsehandler.Build(http.StatusOK, "", nil))
}

func (s *Service) ListMyActuator(c *gin.Context) {
	actuators, err := s.actuatorClient.ListActuator(nil, actuatorclient.ListActuatorQuery{
		Creater: token.GetUserName(c),
	})
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("list my actuator error: %v", err), nil))
		return
	}

	var result []apistructs.Actuator
	for _, dbActuator := range actuators {
		value, err := dbActuator.ToApiStructs()
		if err != nil {
			c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("actuator %v to apiStruct error: %v", dbActuator.Name, err), nil))
			return
		}
		result = append(result, value)
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", result))
}
