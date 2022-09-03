package flowmanager

import (
	"eventops/internal/core/client/eventclient"
	"fmt"
	"sync"
)

type Cache struct {
	lock          sync.Mutex
	clientManager *clientManager

	Events       map[uint64]*eventclient.Event
	EventTrigger map[uint64]*eventclient.EventTrigger
}

func newCache(clientManager *clientManager) *Cache {
	return &Cache{
		clientManager: clientManager,
		Events:        map[uint64]*eventclient.Event{},
		EventTrigger:  map[uint64]*eventclient.EventTrigger{},
	}
}

func (c *Cache) GetAndSetEvent(id uint64) (*eventclient.Event, error) {
	c.lock.Lock()
	cacheEvent := c.Events[id]
	c.lock.Unlock()
	if cacheEvent != nil {
		return cacheEvent, nil
	}

	event, err := c.clientManager.eventTriggerClient.GetEventById(nil, id)
	if err != nil {
		return nil, err
	}
	c.lock.Lock()
	c.Events[id] = event
	c.lock.Unlock()
	return event, nil
}

func (c *Cache) GetAndSetEventTrigger(id uint64) (*eventclient.EventTrigger, error) {
	c.lock.Lock()
	cacheEventTrigger := c.EventTrigger[id]
	c.lock.Unlock()

	if cacheEventTrigger != nil {
		return cacheEventTrigger, nil
	}

	eventTrigger, find, err := c.clientManager.eventTriggerClient.GetEventTrigger(nil, id)
	if err != nil {
		return nil, err
	}
	if !find {
		return nil, fmt.Errorf("not find id: %v event trigger", id)
	}

	c.lock.Lock()
	c.EventTrigger[id] = eventTrigger
	c.lock.Unlock()
	return eventTrigger, nil
}
