package protocol

type MessageType string

const (
	TypeTelemetry MessageType = "telemtry"
	TypeHearbeat  MessageType = "heartbeat"
)

type Message struct {
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Body      any         `json:"body"`
}
