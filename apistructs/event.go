package apistructs

type EventStatus string

const EventCreatedStatus EventStatus = "created"
const EventProcessingStatus EventStatus = "processing"
const EventProcessedStatus EventStatus = "processed"

type Value string

type FileValueType string

type FileValue struct {
	Value string        `json:"value"`
	Type  FileValueType `json:"type"`
}

type Event struct {
	Name         string               `json:"name"`
	Version      string               `json:"version"`
	Values       map[string]Value     `json:"values"`
	Files        map[string]FileValue `json:"files"`
	Labels       map[string]string    `json:"labels"`
	Timestamp    int64                `json:"timestamp"`
	SupportUsers []string             `json:"users"`
}
