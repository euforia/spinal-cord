package io

import (
	"encoding/json"
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	revent "github.com/euforia/spinal-cord/revent/v2"
	zmq "github.com/pebbe/zmq3"
)

type ZmqPublisher struct {
	sockType string
	Sock     *zmq.Socket
	Logger   *logging.Logger
}

func NewZmqPublisher(cfg config.IOConfig, logger *logging.Logger) (*ZmqPublisher, error) {
	//func NewZmqPublisher(listenAddr string, logger *logging.Logger) *ZmqPublisher {
	var (
		err error
		z   ZmqPublisher = ZmqPublisher{Logger: logger}
		ok  bool
	)

	z.sockType, ok = cfg.Config["type"].(string)
	if !ok {
		return &z, fmt.Errorf("invalid type: %s", cfg.Config["type"])
	}
	switch z.sockType {
	case "PUB":
		if z.Sock, err = zmq.NewSocket(zmq.PUB); err != nil {
			return &z, err
		}
		break
	default:
		return &z, fmt.Errorf("Invalid zeromq type: %s", z.sockType)
	}

	listenAddr := fmt.Sprintf("tcp://*:%d", cfg.Port)

	if err = z.Sock.Bind(listenAddr); err != nil {
		return &z, fmt.Errorf("Failed to start spinal output: %s %v\n", listenAddr, err)
	}

	logger.Warning.Printf("Spinal output listening: %s %s\n", z.sockType, listenAddr)
	return &z, nil
}

func (p *ZmqPublisher) Start(pchan chan revent.Event) {
	go func(ch chan revent.Event) {
		p.Logger.Warning.Printf("Started spinal output: %s\n", p.sockType)
		for {
			p.Logger.Trace.Println("Waiting for events...")
			evt := <-ch

			b, err := json.Marshal(&evt)
			if err != nil {
				p.Logger.Warning.Printf("Could not jsonify: %s\n", err)
				continue
			}
			p.Logger.Debug.Printf("Publishing: %s\n", b)
			p.Sock.SendBytes(b, 0)
		}
	}(pchan)
}

func (p *ZmqPublisher) Stop() {

}
