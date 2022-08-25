package eventprocess

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"tiggerops/apistructs"
	"tiggerops/conf"
	"tiggerops/internal/event/client/eventclient"
	"tiggerops/internal/register/client/triggerdefinitionclient"
	"tiggerops/pkg/schema/event"
	"time"
)

type Process struct {
	Buffer  chan eventclient.Event
	WorkNum int64

	eventDbClient           *eventclient.Client
	triggerDefinitionClient *triggerdefinitionclient.Client
	dbClient                *gorm.DB
	ctx                     context.Context

	Cache gcache.Cache
}

func NewProcess(dbClient *gorm.DB, ctx context.Context) *Process {
	return &Process{
		Buffer:                  make(chan eventclient.Event, conf.GetEvent().Process.BufferSize),
		WorkNum:                 conf.GetEvent().Process.WorkNum,
		eventDbClient:           eventclient.NewEventClient(dbClient),
		triggerDefinitionClient: triggerdefinitionclient.NewTriggerDefinitionClient(dbClient),
		dbClient:                dbClient,
		ctx:                     ctx,
		Cache:                   gcache.New(conf.GetEvent().Process.TriggerCacheSize).LRU().Build(),
	}
}

func (p *Process) MakeCacheKey(eventName string, eventVersion string, eventCreater string) string {
	return fmt.Sprintf("%s-%s-%s", eventName, eventVersion, eventCreater)
}

func (p *Process) DeleteTriggerCache(key string) bool {
	return p.Cache.Remove(key)
}

func (p *Process) setTriggerCache(key string, triggerList []Trigger) error {
	return p.Cache.Set(key, triggerList)
}

func (p *Process) getTriggerCache(key string) []Trigger {
	values, err := p.Cache.Get(key)
	if err != nil {
		return nil
	}
	return values.([]Trigger)
}

func (p *Process) LoopLoadEvent() {
	go func() {
		for {
			events, err := p.eventDbClient.ListEvent(nil, eventclient.EventQuery{
				Statues: []apistructs.EventStatus{
					apistructs.EventCreatedStatus,
					apistructs.EventProcessingStatus,
				},
			})
			if err != nil {
				logrus.Errorf("[event process] list event error: %v", err)
				time.Sleep(time.Duration(conf.GetEvent().Process.LoopLoadEventInterval) * time.Second)
				continue
			}

			for _, e := range events {
				if e.Status == apistructs.EventCreatedStatus {
					p.AddToProcess(e)
				}
				if e.Status == apistructs.EventProcessingStatus && time.Now().Unix()-e.UpdatedAt.Unix() > conf.GetEvent().Process.ProcessingOverTime {
					err := p.eventDbClient.UpdateEventStatus(nil, e.Id, e.Status, apistructs.EventCreatedStatus, "")
					if err != nil {
						logrus.Error("[event process] update event: %s status to %s status error: %v", e.Status, apistructs.EventCreatedStatus, err)
						continue
					}
					e.Status = apistructs.EventCreatedStatus
					p.AddToProcess(e)
				}
			}
			time.Sleep(time.Duration(conf.GetEvent().Process.LoopLoadEventInterval) * time.Second)
		}
	}()
}

func (p *Process) AddToProcess(event eventclient.Event) {
	go func() {
		p.Buffer <- event
	}()
}

func (p *Process) Run() {
	for i := 0; i < int(p.WorkNum); i++ {
		go func() {
			for {
				select {
				case <-p.ctx.Done():
				case bufferEvent, ok := <-p.Buffer:
					if !ok {
						continue
					}
					if bufferEvent.Status != apistructs.EventCreatedStatus {
						continue
					}
					err := p.process(bufferEvent)
					if err != nil {
						logrus.Error("process event name %v version %v creater %v error: %v", bufferEvent.Name, bufferEvent.Version, bufferEvent.Creater, err)
					}
				}
			}
		}()
	}
}

type Trigger struct {
	event.Trigger
	Creater string `json:"creater"`
}

func (p *Process) getAndSetTriggerCaches(dbEvent *eventclient.Event) ([]Trigger, error) {
	triggers := p.getTriggerCache(p.MakeCacheKey(dbEvent.Name, dbEvent.Version, dbEvent.Creater))
	if triggers == nil {
		dbTriggers, err := p.triggerDefinitionClient.ListEventTriggerDefinition(nil, triggerdefinitionclient.ListEventTriggerDefinitionQuery{
			EventCreater: dbEvent.Creater,
			EventVersion: dbEvent.Version,
			EventName:    dbEvent.Name,
		})
		if err != nil {
			return nil, err
		}
		var cachesTriggers []Trigger
		for _, dbTrigger := range dbTriggers {
			var eventTrigger event.Trigger
			err := yaml.Unmarshal([]byte(dbTrigger.Content), &eventTrigger)
			if err != nil {
				// todo 有可能导致这个 trigger 永远无法使用，待观察和测试
				logrus.Error("failed to Unmarshal trigger content: %v", dbTrigger.Content)
				continue
			}

			var trigger Trigger
			trigger.Trigger = eventTrigger
			trigger.Creater = dbTrigger.Creater
			cachesTriggers = append(cachesTriggers, trigger)
		}

		err = p.setTriggerCache(p.MakeCacheKey(dbEvent.Name, dbEvent.Version, dbEvent.Creater), cachesTriggers)
		if err != nil {
			return nil, err
		}
		triggers = cachesTriggers
	}
	return triggers, nil
}

func (p *Process) process(processEvent eventclient.Event) error {
	dbEvent, err := p.eventDbClient.GetEventById(nil, processEvent.Id)
	if err != nil {
		return err
	}
	if dbEvent.Status != apistructs.EventCreatedStatus {
		return nil
	}

	if err := p.eventDbClient.UpdateEventStatus(nil, dbEvent.Id, dbEvent.Status, apistructs.EventProcessingStatus, ""); err != nil {
		return err
	}

	var eventInfo apistructs.Event
	if err := json.Unmarshal([]byte(dbEvent.Content), &eventInfo); err != nil {
		return p.eventDbClient.UpdateEventStatus(nil, dbEvent.Id, apistructs.EventProcessingStatus,
			apistructs.EventProcessFailedStatus, fmt.Sprintf("unmarshal event content error %v", err))
	}

	triggers, err := p.getAndSetTriggerCaches(dbEvent)
	if err != nil {
		return p.eventDbClient.UpdateEventStatus(nil, dbEvent.Id, apistructs.EventProcessingStatus,
			apistructs.EventProcessFailedStatus, fmt.Sprintf("get event triggers definition error %v", err))
	}
	if triggers == nil {
		return p.eventDbClient.UpdateEventStatus(nil, dbEvent.Id, apistructs.EventProcessingStatus, apistructs.EventProcessedStatus, "not find triggers definition")
	}

	var eventTriggers []eventclient.EventTrigger
	for _, trigger := range triggers {
		// filter event not support user trigger definition
		var inSupportUser = false
		for _, supportUser := range eventInfo.SupportUsers {
			if trigger.Creater == supportUser {
				inSupportUser = true
				break
			}
		}
		if len(eventInfo.SupportUsers) > 0 && !inSupportUser {
			continue
		}

		for index := range trigger.Pipelines {
			trigger.Pipelines[index].Filters = append(trigger.Pipelines[index].Filters, trigger.Filters...)
		}

		for _, pipeline := range trigger.Pipelines {
			pass := checkPass(pipeline, dbEvent.Content)

			var eventTrigger = eventclient.EventTrigger{
				EventName:      dbEvent.Name,
				EventCreater:   dbEvent.Creater,
				EventVersion:   dbEvent.Version,
				EventTime:      dbEvent.CreatedAt,
				EventId:        dbEvent.Id,
				TriggerName:    trigger.Name,
				TriggerCreater: trigger.Creater,
				TriggerTime:    time.Now(),
				PipelineImage:  pipeline.Image,
			}
			if pass {
				eventTrigger.Status = apistructs.PassEventTriggerStatus
				eventTrigger.Message = ""
			} else {
				eventTrigger.Status = apistructs.UnPassEventTriggerStatus
				eventTrigger.Message = fmt.Sprintf("pipeline %v filter not match", pipeline.Image)
			}
			eventTriggers = append(eventTriggers, eventTrigger)
		}
	}

	return p.dbClient.Transaction(func(tx *gorm.DB) error {
		if err := p.eventDbClient.UpdateEventStatus(tx, dbEvent.Id, apistructs.EventProcessingStatus, apistructs.EventProcessedStatus, ""); err != nil {
			return err
		}
		if len(eventTriggers) > 0 {
			if err := p.eventDbClient.BatchCreateEventTrigger(tx, eventTriggers); err != nil {
				return err
			}
		}
		return nil
	})
}

func checkPass(pipeline event.TriggerPipeline, eventContent string) bool {
	var pass = true

	for _, filter := range pipeline.Filters {
		value := gjson.Get(eventContent, filter.Expr).String()
		for _, match := range filter.Matches {
			if match == value {
				pass = true
				break
			}
		}
	}
	return pass
}