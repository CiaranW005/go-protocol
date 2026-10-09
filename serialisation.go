package protocol

import (
	"encoding/json"
	"fmt"
)

func SerialisePacket(packet any) ([]byte, error) {
	return json.Marshal(packet)
}

func DeserialisePacket(raw []byte) (PacketVariant, error) {
	var header struct {
		Type   *MessageType `json:"type"`
		Status *AckStatus   `json:"status"`
	}

	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, err
	}

	if header.Status != nil {
		var ack Acknowledgement

		if err := json.Unmarshal(raw, &ack); err != nil {
			return nil, err
		}

		return ack, nil
	}

	if header.Type != nil {
		var msg Message

		if err := json.Unmarshal(raw, &msg); err != nil {
			return nil, err
		}

		return msg, nil
	}

	return nil, fmt.Errorf("Packet contained neither type nor status")
}
