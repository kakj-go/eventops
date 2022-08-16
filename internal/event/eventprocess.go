package event

import (
	"context"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/goccy/go-json"
	"github.com/ohler55/ojg/jp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"tiggerops/apistructs"
	"tiggerops/conf"
	"tiggerops/internal/event/client/eventclient"
	"tiggerops/internal/register/client/triggerdefinitionclient"
	"tiggerops/pkg/schema/event"
	"time"
)

const processOverTime = 3600

type Process struct {
	Buffer  chan eventclient.Event
	WorkNum int64

	eventDbClient           *eventclient.Client
	triggerDefinitionClient *triggerdefinitionclient.Client
	dbClient                *gorm.DB
	ctx                     context.Context

	Cache gcache.Cache
}

func NewProcess(eventDbClient *eventclient.Client, triggerDefinitionClient *triggerdefinitionclient.Client, dbClient *gorm.DB, ctx context.Context) Process {
	return Process{
		Buffer:                  make(chan eventclient.Event, conf.GetEvent().Process.BufferSize),
		WorkNum:                 conf.GetEvent().Process.WorkNum,
		eventDbClient:           eventDbClient,
		triggerDefinitionClient: triggerDefinitionClient,
		dbClient:                dbClient,
		ctx:                     ctx,
		Cache:                   gcache.New(conf.GetEvent().Process.CacheSize).LRU().Build(),
	}
}

func (p *Process) MakeCacheKey(eventName string, eventVersion string, eventCreater string) string {
	return fmt.Sprintf("%s-%s-%s", eventName, eventVersion, eventCreater)
}

func (p *Process) SetTriggerCache(key string, triggerList []Trigger) error {
	return p.Cache.Set(key, triggerList)
}

func (p *Process) GetTriggerCache(key string) ([]Trigger, error) {
	values, err := p.Cache.Get(key)
	if err != nil {
		return nil, err
	}
	return values.([]Trigger), nil
}

func (p *Process) loopLoadEvent() {
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
				if e.Status == apistructs.EventProcessingStatus && e.UpdatedAt.Unix()-time.Now().Unix() > processOverTime {
					err := p.eventDbClient.UpdateEventStatus(nil, e.Id, e.Status, apistructs.EventCreatedStatus, "")
					if err != nil {
						logrus.Error("[event process] update event: %s status to %s status error: %v", e.Status, apistructs.EventCreatedStatus, err)
						continue
					}
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
					if bufferEvent.Status != apistructs.EventProcessingStatus {
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

func (p *Process) getAndSetCaches(dbEvent *eventclient.Event) ([]Trigger, error) {
	triggers, err := p.GetTriggerCache(p.MakeCacheKey(dbEvent.Name, dbEvent.Version, dbEvent.Creater))
	if err != nil {
		return nil, err
	}
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
			var trigger Trigger
			err := json.Unmarshal([]byte(dbTrigger.Content), &trigger)
			if err != nil {
				// todo 有可能导致这个 trigger 永远无法使用，待观察和测试
				logrus.Error("failed to Unmarshal trigger content: %v", dbTrigger.Content)
				continue
			}
			trigger.Creater = dbTrigger.Creater
			cachesTriggers = append(cachesTriggers, trigger)
		}

		err = p.SetTriggerCache(p.MakeCacheKey(dbEvent.Name, dbEvent.Version, dbEvent.Creater), cachesTriggers)
		if err != nil {
			return nil, err
		}
		triggers = cachesTriggers
	}
	return triggers, nil
}

func (p *Process) process(processEvent eventclient.Event) error {
	dbEvent, err := p.eventDbClient.GetEvent(nil, processEvent.Name, processEvent.Version, processEvent.Creater)
	if err != nil {
		return err
	}
	if dbEvent.Status != apistructs.EventProcessingStatus {
		return nil
	}

	triggers, err := p.getAndSetCaches(dbEvent)
	if err != nil {
		return err
	}
	if triggers == nil {
		return p.eventDbClient.UpdateEventStatus(nil, dbEvent.Id, dbEvent.Status, apistructs.EventProcessedStatus, "")
	}

	var eventTriggers []eventclient.EventTrigger
	for _, trigger := range triggers {
		for _, pipeline := range trigger.Pipelines {
			pipeline.Filters = append(pipeline.Filters, trigger.Filters...)
		}

		for _, pipeline := range trigger.Pipelines {
			pass, msg := checkPass(pipeline)

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
				eventTrigger.Message = msg
			}
			eventTriggers = append(eventTriggers, eventTrigger)
		}
	}

	return p.dbClient.Transaction(func(tx *gorm.DB) error {
		if err := p.eventDbClient.UpdateEventStatus(tx, dbEvent.Id, dbEvent.Status, apistructs.EventProcessedStatus, ""); err != nil {
			return err
		}
		if err := p.eventDbClient.BatchCreateEventTrigger(tx, eventTriggers); err != nil {
			return err
		}
		return nil
	})
}

func checkPass(pipeline event.TriggerPipeline) (bool, string) {
	var pass = true
	var msg string

	for _, filter := range pipeline.Filters {
		value, err := jp.ParseString(filter.Expr)
		if err != nil {
			pass = false
			msg = fmt.Sprintf("filter: %v parseString error: %v", filter.Expr, err.Error())
			break
		}
		parseString := value.String()
		for _, match := range filter.Matches {
			if match == parseString {
				pass = true
				break
			}
		}
	}
	return pass, msg
}
