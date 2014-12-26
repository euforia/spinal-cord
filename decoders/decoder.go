package decoders

import (
	"encoding/json"
	"fmt"
	revent "github.com/euforia/spinal-cord/revent/v2"
)

type IDecoder interface {
	Decode([]byte) (*revent.Event, error)
}

type JSONDecoder struct {
	//DefaultNamespace string
}

func NewJSONDecoder() error {
	return nil
}

func (j *JSONDecoder) Decode(b []byte) (*revent.Event, error) {
	var event revent.Event
	err := json.Unmarshal(b, &event)
	if err != nil {
		return &event, err
	}
	/*
		if event.Namespace == "" {
			event.Namespace = j.DefaultNamespace
		}
	*/
	return &event, nil
}

func LoadDecoderByName(name string) (IDecoder, error) {
	switch name {
	case "OpenStackAMQPDecoder":
		return &OpenStackAMQPDecoder{}, nil
	default:
		break
	}
	return nil, fmt.Errorf("Decoder not supported: %s", name)
}
