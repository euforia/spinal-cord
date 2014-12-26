package io

import (
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	revent "github.com/euforia/spinal-cord/revent/v2"
)

type SpinalInputType string
type SpinalOutputType string

const (
	INPUT_TYPE_ZPULL SpinalInputType = "zpull"
	INPUT_TYPE_ZREP  SpinalInputType = "zrep"
	INPUT_TYPE_HTTP  SpinalInputType = "http"
)

const (
	OUTPUT_TYPE_ZPUB SpinalOutputType = "zpub"
)

const IO_STARTUP_SUMMARY string = `
	Running Configuration:

		Inputs : %d/%d loaded
`

type ISpinalInputOuput interface {
	Start(chan revent.Event)
	Stop()
}

type IOManager struct {
	InputsMap map[string]ISpinalInputOuput
	//OutputsMap map[string]ISpinalInputOuput
	Publisher ISpinalInputOuput
	CommChan  chan revent.Event
	logger    *logging.Logger
}

func NewIOManager(logger *logging.Logger) *IOManager {
	return &IOManager{
		InputsMap: make(map[string]ISpinalInputOuput),
		CommChan:  make(chan revent.Event),
		logger:    logger}
}

func (i *IOManager) LoadIO(cfg *config.SpinalCordConfig, printSummary bool) {
	i.logger.Info.Printf("Found %d spinal inputs.\n", len(cfg.Inputs))
	i.LoadInputs(cfg)
	i.LoadPublisher(cfg)

	if printSummary {
		fmt.Printf(IO_STARTUP_SUMMARY, len(i.InputsMap), len(cfg.Inputs))
		fmt.Println()
	}
}

func (i *IOManager) LoadInputs(cfg *config.SpinalCordConfig) {
	for k, iCfg := range cfg.Inputs {
		if iCfg.Enabled {
			z, err := LoadSpinalInput(iCfg, i.logger)
			if err != nil {
				i.logger.Error.Printf("Could not start input '%s' %s\n", iCfg.Type, err)
				continue
			}
			z.Start(i.CommChan)
			i.InputsMap[k] = z
		} else {
			i.logger.Debug.Printf("SKIPPING disabled INPUT: %s\n", iCfg.Type)
		}
	}
}

func (i *IOManager) LoadPublisher(cfg *config.SpinalCordConfig) error {
	switch cfg.Core.Publisher.Type {
	case "zeromq":
		z, err := NewZmqPublisher(cfg.Core.Publisher, i.logger)
		if err != nil {
			return err
		}
		z.Start(i.CommChan)
		i.Publisher = z
		return nil
	default:
		return fmt.Errorf("Input type not supported: %s", cfg.Core.Publisher.Type)
	}
}

/*
func LoadPublisher(cfg config.IOConfig, logger *logging.Logger) (ISpinalInputOuput, error) {

}
*/
/*
func LoadSpinalOutput(cfg config.IOConfig, logger *logging.Logger) (ISpinalInputOuput, error) {
	switch cfg.Type {
	case "zeromq":
		return NewZmqSpinalOutput(cfg, logger)
	default:
		return &ZmqSpinalOutput{}, fmt.Errorf("Input type not supported: %s", cfg.Type)
	}
}
*/
func LoadSpinalInput(cfg config.IOConfig, logger *logging.Logger) (ISpinalInputOuput, error) {
	switch cfg.Type {
	case "zeromq":
		return NewZmqSpinalInput(cfg, logger)
		//return ZInputService{}, fmt.Errorf("TBI")
	case "http":
		return NewHttpSpinalInput(cfg, logger)
	default:
		return &ZmqSpinalInput{}, fmt.Errorf("Input type not supported: %s", cfg.Type)
	}
}
