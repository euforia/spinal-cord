package synapse

import (
	"github.com/euforia/spinal-cord/config"
	"testing"
)

var testSynapseConfig = config.SynapseConfig{
	SpinalUri:  "tcp://127.0.0.1:45454",
	Type:       "zpush",
	TypeConfig: config.ZmqSynapseConfig{"PUSH"},
}

func Test_LoadSynapse(t *testing.T) {

}
