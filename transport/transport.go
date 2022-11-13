package transport

import (
	"encoding/json"
	"errors"
)

// This package handles message manipulation between entities within the system
// They should be used as following:
//
// Incoming message from websocket client (Client) uses DecodeIncomingServiceMessage
// to allow reading of the

// OutgoingServiceMessage Outgoing message from a service to relay to client
type OutgoingServiceMessage struct {
	SessionId string `json:"session_id"`
	Topic     string `json:"topic"`
	Reply     string `json:"reply"`
}

// OutgoingClientMessage Message being sent back to client after service reply
// Differs to the OutgoingServiceMessage by not carrying a session id as the
// client is not trusted to hold the session id.
type OutgoingClientMessage struct {
	Topic string `json:"topic"`
	Reply string `json:"reply"`
}

// DecodeServerMessage Convert message to a struct for usage
func DecodeServerMessage(message []byte) (*OutgoingServiceMessage, error) {
	reply := &OutgoingServiceMessage{}
	err := json.Unmarshal(message, reply)
	if err != nil {
		return nil, errors.New("unable to decode/unmarshal server message")
	}

	return reply, nil
}

// Encode OutgoingServiceMessage to bytes for transmission
func (message OutgoingServiceMessage) Encode() ([]byte, error) {
	msg, err := json.Marshal(message)

	if err != nil {
		return nil, errors.New("unable to encode/marshal OutgoingServiceMessage")
	}

	return msg, nil
}

// ToOutgoingClientMessage converts service reply to client return message
func (message OutgoingServiceMessage) ToOutgoingClientMessage() *OutgoingClientMessage {
	return &OutgoingClientMessage{
		Topic: message.Topic,
		Reply: message.Reply,
	}
}

// Encode converts OutgoingClientMessage to bytes
func (message OutgoingClientMessage) Encode() ([]byte, error) {
	msg, err := json.Marshal(message)

	if err != nil {
		return nil, errors.New("unable to encode/marshal OutgoingClientMessage")
	}

	return msg, nil
}
