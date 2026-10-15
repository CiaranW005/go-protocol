package protocol

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
	Type    MessageType `json:"type"`
	ReplyTo ID          `json:"reply_to,omitempty"`
	Body    any         `json:"body"`
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
