package core

import "time"

type AckStatus string

const (
	AckOk             AckStatus = "ok"
	AckError          AckStatus = "error"
	AckInvalidCommand AckStatus = "invalid_command"
	AckBusy           AckStatus = "busy"
)

type Acknowledgement struct {
	Packet
	Status AckStatus `json:"status"`
	Error  string    `json:"error,omitempty"`
}

func (Acknowledgement) isPacket() {}

func NewAcknowledgement(
	message_id ID,
	status AckStatus,
	err ...string,
) Acknowledgement {
	errMsg := ""

	if len(err) > 0 {
		errMsg = err[0]
	}

	return Acknowledgement{
		Packet: Packet{
			ID:        message_id,
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Status: status,
		Error:  errMsg,
	}
}
