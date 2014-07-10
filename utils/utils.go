package utils

import (
	"../logger"
	"encoding/json"
	"strings"
)

type Message struct {
	Command string
	Payload string
}

func (message *Message) Pack() string {
	b, err := json.Marshal(&message)
	if err != nil {
		logger.Error.Println(err)
	}

	s := strings.Replace(string(b), "Command", "command", -1)
	return strings.Replace(s, "Payload", "payload", -1)
}

func Unpack(jsonBlob []byte) Message {
	var message Message
	err := json.Unmarshal(jsonBlob, &message)
	if err != nil {
		logger.Error.Println(err)
	}

	return message
}
