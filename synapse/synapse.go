package synapse

import (
	"fmt"
	"github.com/euforia/spinal-cord/config"
	revent "github.com/euforia/spinal-cord/revent/v2"
)

type SynapseType string

const (
	SYN_TYPE_ZPUSH SynapseType = "zpush"
	SYN_TYPE_ZREQ  SynapseType = "zreq"
)

type ISynapse interface {
	Fire(*revent.Event) error
}

func LoadSynapse(synCfg config.SynapseConfig) (ISynapse, error) {
	switch SynapseType(synCfg.Type) {
	case SYN_TYPE_ZPUSH:
		return NewZMQSynapse(synCfg)
	case SYN_TYPE_ZREQ:
		return NewZMQSynapse(synCfg)
	default:
		return nil, fmt.Errorf("Synapse type not supported: %s", synCfg.Type)
	}
}
