package protocol

import (
	"github.com/CiaranW005/go-protocol/internal/core"
	"github.com/CiaranW005/go-protocol/tracker"
)

// Expose packet types and trackers through one package so callers do not
// need to import the implementation packages separately.
const (
	TypeWake      = core.TypeWake
	TypeReset     = core.TypeReset
	TypePowerOff  = core.TypePowerOff
	TypeTelemetry = core.TypeTelemetry
	TypeHeartbeat = core.TypeHeartbeat
	TypeUnknown   = core.TypeUnknown

	AckOk             = core.AckOk
	AckError          = core.AckError
	AckInvalidCommand = core.AckInvalidCommand
	AckBusy           = core.AckBusy
)

type (
	ID              = core.ID
	Message         = core.Message
	MessageType     = core.MessageType
	Acknowledgement = core.Acknowledgement
	PacketVariant   = core.PacketVariant
	CommandResult   = core.CommandResult

	AckTracker       = tracker.AckTracker
	ResponseTracker  = tracker.ResponseTracker
	RetryPolicy      = tracker.RetryPolicy
	TrackerCallbacks = tracker.TrackerCallbacks
)

var (
	NewMessage         = core.NewMessage
	NewAcknowledgement = core.NewAcknowledgement
	SerialisePacket    = core.SerialisePacket
	DeserialisePacket  = core.DeserialisePacket

	NewAckTracker      = tracker.NewAckTracker
	NewResponseTracker = tracker.NewResponseTracker

	ErrResponseTimeout = tracker.ErrResponseTimeout
)
