package decoders

import (
	"encoding/json"
	"fmt"
	reventV1 "github.com/euforia/spinal-cord/revent/v1"
	revent "github.com/euforia/spinal-cord/revent/v2"
	"time"
)

/* must use the literal string as mentioned in the docs for time layout */
const OPENSTACK_TIMESTAMP_FORMAT string = "2006-01-02 15:04:05.000000"

type OpenStackAMQPDecoder struct{}

func (o *OpenStackAMQPDecoder) Decode(b []byte) (*revent.Event, error) {
	var (
		eventV1 reventV1.Event /* old schema containing openstack amqp format */
		event   revent.Event   = revent.Event{}
	)
	err := json.Unmarshal(b, &eventV1)
	if err != nil {
		return &event, err
	}
	event.Namespace = eventV1.Namespace
	event.Type = eventV1.Type
	event.Data = eventV1.Payload

	timeStr, ok := eventV1.Timestamp.(string)
	if !ok {
		return &event, fmt.Errorf("timestamp not a string: %s", eventV1.Timestamp)
	}

	parsedTime, err := time.Parse(OPENSTACK_TIMESTAMP_FORMAT, timeStr)
	if err != nil {
		return &event, err
	}
	event.Timestamp = float64(parsedTime.UnixNano()) / 1000000000

	return &event, nil
}
