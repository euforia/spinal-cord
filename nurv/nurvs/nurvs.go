package nurvs

import (
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/decoders"
	"github.com/euforia/spinal-cord/logging"
	"github.com/euforia/spinal-cord/synapse"
)

type NurvType string

const (
	NURV_TYPE_AMQP NurvType = "amqp"
)

type INurv interface {
	/* decoder, synapse */
	Init(decoders.IDecoder, synapse.ISynapse) error
	/* must be non-blocking */
	Start() error
	Stop() error
}

func LoadNurv(cfg *config.NurvConfig, logger *logging.Logger) (INurv, error) {
	switch NurvType(cfg.Type) {
	case NURV_TYPE_AMQP:
		var (
			amqpNurv *AMQPNurv
			err      error = nil
		)
		amqpNurv, err = NewAMQPNurv(cfg, logger)
		if err != nil {
			return amqpNurv, err
		}

		dcd, err := decoders.LoadDecoderByName(cfg.Decoder)
		if err != nil {
			return amqpNurv, err
		}

		syn, err := synapse.LoadSynapse(cfg.Synapse)
		if err != nil {
			return amqpNurv, err
		}

		amqpNurv.Init(dcd, syn)
		return amqpNurv, nil
	default:
		return nil, fmt.Errorf("type not supported: %s", cfg.Type)
	}
}
