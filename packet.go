package protocol

type MessageType string

const (
	TypeReset     MessageType = "reset"
	TypePowerOff  MessageType = "power_off"
	TypeTelemetry MessageType = "telemetry"
	TypeHeartbeat MessageType = "heartbeat"
	TypeUnknown   MessageType = "unknown"
)

type AckStatus string

const (
	AckOk             AckStatus = "ok"
	AckError          AckStatus = "error"
	AckInvalidCommand AckStatus = "invalid_command"
	AckBusy           AckStatus = "busy"
)

type Packet struct {
	Timestamp uint64 `json:"timestamp"`
}

type PacketVariant interface {
	isPacket()
}

type Message struct {
	Packet
	Type MessageType `json:"type"`
	Body any         `json:"body"`
}

func (Message) isPacket() {}

func NewMessage(timestamp uint64, t MessageType, body any) Message {
	return Message{
		Packet: Packet{Timestamp: timestamp},
		Type:   t,
		Body:   body,
	}
}

type Acknowledgement struct {
	Packet
	Status AckStatus `json:"status"`
	Error  string    `json:"error"`
}

func (Acknowledgement) isPacket() {}

func NewAcknowledgement(timestamp uint64, status AckStatus, err string) Acknowledgement {
	return Acknowledgement{
		Packet: Packet{Timestamp: timestamp},
		Status: status,
		Error:  err,
	}
}
