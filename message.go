package socket

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
)

type Message struct {
	TypeMessage string					`json:"type"`
	Data 		map[string]interface{}	`json:"data"`
}

func NewMessage(msg []byte) (*Message, error) {
	newMessage := Message{}
	if err := json.Unmarshal(msg, &newMessage); err != nil {
		return nil, err
	}
	return &newMessage, nil
}

func BindData(data map[string]interface{}, obj interface{}, tagNames ...string) error {

	decoderConfig := &mapstructure.DecoderConfig{
		Result: obj,
		TagName: "json",
	}

	if len(tagNames) > 0 {
		decoderConfig.TagName = tagNames[0]
	}

	decoder,err  := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(data)
}
