package synapse

import (
	"github.com/euforia/spinal-cord/config"
	"testing"
)

var testSynapseConfig = config.SynapseConfig{
	URI:        "tcp://127.0.0.1:45454",
	Type:       "zpush",
	TypeConfig: config.ZmqSynapseConfig{"PUSH"},
}

var testRecvSynapseConfig = config.SynapseConfig{
	URI:        "tcp://127.0.0.1:55000",
	Type:       "zsub",
	TypeConfig: config.RecvSynapseConfig{Type: "SUB", Subscriptions: []string{"a", "b"}},
}

func Test_LoadSynapse(t *testing.T) {

}
