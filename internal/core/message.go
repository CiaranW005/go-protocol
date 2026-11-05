package core

import "time"

type MessageType string

type CommandResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

const (
	TypeWake          MessageType = "wake"
	TypeReset         MessageType = "reset"
	TypePowerOff      MessageType = "power_off"
	TypeTelemetry     MessageType = "telemetry"
	TypeHeartbeat     MessageType = "heartbeat"
	TypeCommandResult MessageType = "command_results"
	TypeUnknown       MessageType = "unknown"
)

type Message struct {
	Packet
	Type MessageType `json:"type"`
	// ReplyTo identifies the original request when this message is a response.
	// Leave empty for a new request.
	ReplyTo ID `json:"reply_to,omitempty"`
	// Body must be JSON-encodable. After decoding, JSON objects become
	// map[string]any rather than their original Go struct type.
	Body any `json:"body"`
}

func (Message) isPacket() {}

func NewMessage(t MessageType, body any) Message {
	return Message{
		Packet: Packet{
			ID:        NewID(),
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Type: t,
		Body: body,
	}
}
