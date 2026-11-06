# BMC Go Protocol

This module defines the JSON packets and message-tracking utilities shared by the [BMC](../README.md) Oracle server and PC agent. The ESP32 implements matching packet structures in C++.

The root `protocol` package exposes the core types, constructors, encoding functions, and trackers. Packet implementations live in `internal/core`; acknowledgement and response tracking live in `tracker`.

## Packet format

A message has its own UUID, a timestamp, a command type, a JSON body, and an optional reference to an earlier request:

```json
{
  "id": "7a3ef547-1645-4650-99ef-6b3b271a8b22",
  "timestamp": 1800000000000,
  "type": "telemetry",
  "body": {"pc": true},
  "reply_to": "5fe8c943-39bd-49b3-87dc-88baf7ac1428"
}
```

`NewMessage` generates a UUID and a Unix timestamp in milliseconds. Callers set `ReplyTo` when constructing a response. `Body` accepts JSON-encodable values; decoded JSON objects become `map[string]any`, not the original Go struct.

An acknowledgement reuses the acknowledged message's ID:

```json
{
  "id": "7a3ef547-1645-4650-99ef-6b3b271a8b22",
  "timestamp": 1800000000001,
  "status": "ok"
}
```

Acknowledgement statuses are `ok`, `error`, `invalid_command`, and `busy`. An optional `error` field carries details. Common message types include `wake`, `reset`, `power_off`, `telemetry`, and `heartbeat`.

`DeserialisePacket` identifies acknowledgements by a non-null `status` field and messages by a non-null `type` field. Status takes precedence if both are present. It decodes the structure but does not validate command names or require non-empty IDs. Application handlers perform further checks.

Transport framing is outside this module: the Oracle connection uses WebSocket messages, while the PC agent uses newline-delimited JSON over TCP.

## Delivery and completion

`AckTracker` waits for acknowledgement of delivery. `Track` registers a message but does not send it; the caller performs the initial send. `RetryPolicy.Timeout` applies to each attempt, and `MaxRetries` counts sends after the initial one. Zero disables retries while still waiting for acknowledgement.

Any matching acknowledgement stops retries, including a rejection. The `Acked` callback inspects its status. Retry and failure callbacks execute outside the tracker lock so they can safely call tracking code again.

`ResponseTracker` waits for a result message whose `ReplyTo` matches the original request ID. `Track` returns a channel buffered for one result. Resolution, failure, cancellation, or timeout removes the entry and supplies a result. `Cancel` and `Clear` notify waiters with a cancellation error.

Both trackers use a background `Run(ctx)` loop that checks deadlines every 100 milliseconds. Canceling the context stops this loop but does not itself clear pending entries or notify waiters. The generic base tracker's `Cancel` and `Clear` only remove entries; the response tracker adds notification behavior.

## Duplicate detection

`Deduplicator.IsDuplicate` records unseen IDs and returns true for IDs encountered again before their expiry. Duplicate checks do not extend that expiry. It is protected by a mutex.

The current Go implementation retains map entries for expired IDs until those IDs are reused; it does not sweep old entries. Its lifetime controls duplicate recognition, not a bound on memory use.

## Using the module

The BMC parent repository links this module through `go.work`. Use Go 1.26.0. Import the public package as:

```go
import protocol "github.com/CiaranW005/go-protocol"
```

Create and encode a request:

```go
message := protocol.NewMessage(protocol.TypeTelemetry, nil)
data, err := protocol.SerialisePacket(message)
```

For tracked operations, register before sending so a fast acknowledgement or response can be matched. The Oracle server's device manager and the PC agent's connection manager show the complete integration.

## Cross-language compatibility

Changes to JSON field names, message types, status values, or timing conventions must be checked against the ESP32 implementation as well as both Go callers. Currently the Go constructors emit timestamps in Unix milliseconds, while some firmware paths use Unix seconds or device uptime; timestamps are not used to correlate replies, but their meaning is not yet uniform across implementations.
