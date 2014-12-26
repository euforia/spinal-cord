package io

import (
	"encoding/json"
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	revent "github.com/euforia/spinal-cord/revent/v2"
	zmq "github.com/pebbe/zmq3"
	"time"
)

type BasicSock struct {
	Sock   *zmq.Socket
	Logger *logging.Logger
}

func NewBasicSock(sock *zmq.Socket, logger *logging.Logger) BasicSock {
	return BasicSock{sock, logger}
}

type ZmqSpinalInput struct {
	sockType string
	BasicSock
}

func NewZmqSpinalInput(cfg config.IOConfig, logger *logging.Logger) (*ZmqSpinalInput, error) {
	//func NewZmqSpinalInput(sockType string, listenAddr string, logger *logging.Logger) *ZmqSpinalInput {
	var (
		zis   ZmqSpinalInput
		zsock *zmq.Socket
		err   error
	)

	sockType, ok := cfg.Config["type"].(string)
	if !ok {
		return &zis, fmt.Errorf("Invalid input type: %s", cfg.Config["type"])
	}

	switch sockType {
	case "PULL":
		zsock, err = zmq.NewSocket(zmq.PULL)
		break
	case "REP":
		zsock, _ = zmq.NewSocket(zmq.REP)
		break
	default:
		return &zis, fmt.Errorf("Invalid sock type: %s\n", sockType)
	}

	listenAddr := fmt.Sprintf("tcp://*:%d", cfg.Port)
	err = zsock.Bind(listenAddr)
	if err != nil {
		return &zis, fmt.Errorf("%s %v\n", listenAddr, err)
	}
	logger.Warning.Printf("Spinal input listening: %s %s\n", sockType, listenAddr)

	return &ZmqSpinalInput{sockType, BasicSock{zsock, logger}}, nil
}

/* TODO:
 * pull function in input CORE
 */
func (b *ZmqSpinalInput) CheckMessage(message string) (revent.Event, error) {

	var event revent.Event
	err := json.Unmarshal([]byte(message), &event)
	if err != nil {
		return event, err
	}
	if event.Namespace == "" {
		return event, fmt.Errorf("Namespace required!")
	}
	if event.Type == "" {
		return event, fmt.Errorf("Event - 'Type' required!")
	}
	if event.Timestamp <= 0 {
		b.Logger.Trace.Printf("%s - Adding timestamp to: %s\n", b.sockType, message)
		event.Timestamp = float64(time.Now().UnixNano()) / 1000000000
	}
	return event, nil
}

func (b *ZmqSpinalInput) Start(pchan chan revent.Event) {
	go func(ch chan revent.Event) {
		b.Logger.Info.Printf("Started input: %s\n", b.sockType)
		for {
			msg, err := b.Sock.Recv(0)
			if err != nil {
				b.Logger.Error.Println(err)
				continue
			}
			checkedEvt, err := b.CheckMessage(msg)

			b.Logger.Trace.Printf("%s - %s\n", b.sockType, msg)

			ch <- checkedEvt
			if b.sockType == "REP" {
				chkdMsg, err := json.Marshal(&checkedEvt)
				if err != nil {
					b.Logger.Warning.Printf("Could not jsonify: %s\n", checkedEvt)
					continue
				}
				_, err = b.Sock.SendBytes(chkdMsg, 0)
				if err != nil {
					b.Logger.Warning.Printf("Error sending bytes: %s\n", err)
				}
			}
		}
	}(pchan)
}

func (b *ZmqSpinalInput) Stop() {

}
