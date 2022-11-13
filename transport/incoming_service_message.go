package transport

import (
	"encoding/json"
	"errors"
)

// IncomingServiceMessage Incoming message passing orders to a service
// Used by websocket server to pass incoming client messages to a service
// and services sending orders to other services.
type IncomingServiceMessage struct {
	SessionId string        `json:"session_id"` // Client Identifier
	UserId    string        `json:"user_id"`    // USed by services to relate actions to a user
	Topic     string        `json:"topic"`      // The topic channel to send the command on
	Command   string        `json:"body"`       // The command the client wants executed
	Arguments []interface{} `json:"arguments"`  // The arguments the service may need for executing the command
}

// DecodeIncomingServiceMessage Convert message to a struct for usage
func DecodeIncomingServiceMessage(message []byte) (*IncomingServiceMessage, error) {
	reply := &IncomingServiceMessage{}
	err := json.Unmarshal(message, reply)
	if err != nil {
		return nil, errors.New("unable to decode/unmarshal client message")
	}

	return reply, nil
}

// Encode IncomingServiceMessage to bytes for transmission
func (message IncomingServiceMessage) Encode() ([]byte, error) {
	msg, err := json.Marshal(message)

	if err != nil {
		return nil, errors.New("unable to encode/marshal IncomingServiceMessage message")
	}

	return msg, nil
}
