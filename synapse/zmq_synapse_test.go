package synapse

import (
	"fmt"
	//"github.com/euforia/spinal-cord/config"
	revent "github.com/euforia/spinal-cord/revent/v2"
	//zmq "github.com/pebbe/zmq3"
	"testing"
)

//var testSpinalCordUri = "tcp://127.0.0.1:45454"
var testEvent *revent.Event = revent.NewEvent(
	"testing",
	"zmq.test",
	map[string]interface{}{
		"name":        "ZMQSynapse",
		"description": "Testing synapse fire.",
	})

func Test_ZMQSynapse_Fire(t *testing.T) {
	//s, err := NewZMQSynapse(zmq.PUSH, testSpinalCordUri)
	s, err := NewZMQSynapse(testSynapseConfig)
	if err != nil {
		t.Error(fmt.Sprintf("Failed to create synapse: %s %s", testSynapseConfig.SpinalUri, err))
	}

	err = s.Fire(testEvent)
	if err != nil {
		t.Errorf("Failed to fire synapse: %s %s", testSynapseConfig.SpinalUri, err)
	}
}
