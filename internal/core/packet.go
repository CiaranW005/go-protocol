package core

import "github.com/google/uuid"

type ID string

func NewID() ID {
	return ID(uuid.NewString())
}

type Packet struct {
	ID        ID     `json:"id"`
	Timestamp uint64 `json:"timestamp"` // Timestamp is Unix time in milliseconds.
}

func (p Packet) GetID() ID {
	return p.ID
}

type PacketVariant interface {
	isPacket() // The unexported isPacket method limits implementations to types defined in this package.
	GetID() ID
}
