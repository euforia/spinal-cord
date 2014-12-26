package reactor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	"github.com/euforia/spinal-cord/reactor/handler"
	"github.com/euforia/spinal-cord/reactor/task"
	revent "github.com/euforia/spinal-cord/revent/v2"
	zmq "github.com/pebbe/zmq3"
)

type SubReactor struct {
	Aggregator *zmq.Socket
	TaskServer *zmq.Socket

	handlersMgr *handler.HandlersManager

	logger *logging.Logger
}

func NewSubReactor(cfg *config.SpinalCordConfig, logger *logging.Logger) (*SubReactor, error) {
	//func NewSubReactor(aggrUri string, listenAddr string, handlersDir string, logger *logging.Logger) *SubReactor {
	/* listen for published messages */
	var (
		sreactor SubReactor = SubReactor{}
		err      error
	)
	sreactor.handlersMgr = handler.NewHandlersManager(cfg.Core.HandlersDir, logger)
	sreactor.logger = logger

	switch cfg.Reactor.SpinalCord.Type {
	case "SUB":
		sreactor.Aggregator, err = zmq.NewSocket(zmq.SUB)
		if err != nil {
			return &sreactor, err
		}
		sreactor.Aggregator.Connect(cfg.Reactor.SpinalCord.URI)
		if len(cfg.Reactor.SpinalCord.Subscriptions) <= 0 {
			sreactor.Aggregator.SetSubscribe("")
		} else {
			for _, s := range cfg.Reactor.SpinalCord.Subscriptions {
				sreactor.Aggregator.SetSubscribe(s)
			}
		}
		logger.Warning.Printf("Connected to publisher: %s\n", cfg.Reactor.SpinalCord.URI)
		break
	default:
		return &sreactor, fmt.Errorf("type not supported: %s", cfg.Reactor.SpinalCord.Type)
	}

	/* task worker server */
	if sreactor.TaskServer, err = zmq.NewSocket(zmq.PUSH); err != nil {
		return &sreactor, err
	}

	listenAddr := fmt.Sprintf("tcp://*:%d", cfg.Reactor.Port)
	sreactor.TaskServer.Bind(listenAddr)
	logger.Warning.Printf("Task server started: PUSH %s\n", listenAddr)

	return &sreactor, nil
}

func (s *SubReactor) Start(createSamples bool) {
	s.logger.Info.Printf("Create samples: %s", createSamples)

	commChan := make(chan revent.Event)

	go s.startConsumingFromAggregator(commChan, createSamples)

	s.startPushingToWorkers(commChan)
}

func (s *SubReactor) decodeMessage(msg []string) (revent.Event, error) {
	s.logger.Trace.Printf("Attempting to decode: %v\n", msg)
	var event revent.Event
	switch len(msg) {
	case 1:
		err := json.Unmarshal([]byte(msg[0]), &event)
		if err != nil {
			return event, err
		}
		break
	case 2:
		err := json.Unmarshal([]byte(msg[1]), &event)
		if err != nil {
			return event, err
		}
		break
	default:
		return event, errors.New("Invalid message length")
	}
	return event, nil
}

func (s *SubReactor) assembleTask(evtBytes []byte, evtHandler handler.EventHandler) (string, error) {
	hdlr, _ := evtHandler.Handler()

	newtask := task.Task{string(evtBytes), hdlr}
	taskBytes, err := newtask.Serialize()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Couldn't serialize task: %v, reason: %v", newtask, err))
	}
	return string(taskBytes), nil
}

func (s *SubReactor) startPushingToWorkers(ch chan revent.Event) {
	for {
		s.logger.Trace.Println("Waiting for event...")

		event := <-ch
		bytes, err := json.Marshal(&event)
		if err != nil {
			s.logger.Error.Println("Pre-handler:", err)
			continue
		}
		//s.logger.Trace.Printf("Payload: %s\n", string(bytes))

		handlers := s.handlersMgr.GetHandlers(event.Namespace, event.Type)

		s.logger.Info.Printf("Namespace => %s; Event => %s; Handlers => %d\n",
			event.Namespace, event.Type, len(handlers))

		for _, evtHandler := range handlers {
			taskData, err := s.assembleTask(bytes, evtHandler)
			if err != nil {
				s.logger.Error.Println("Failed to assemble task!", err)
				continue
			}
			s.logger.Debug.Printf("Queueing worker task: %s\n", taskData)
			s.TaskServer.Send(taskData, 0)
		}
	}
}

func (s *SubReactor) startConsumingFromAggregator(ch chan revent.Event, createSamples bool) {
	for {
		// Get event from zmq PUB queue //
		msg, err := s.Aggregator.RecvMessage(0)
		if err != nil {
			s.logger.Error.Println("RecvMessage:", err)
			continue
		}
		// Decode event to datastructure //
		zEvent, err := s.decodeMessage(msg)
		if err != nil {
			s.logger.Error.Println("decodeMessage:", err)
			continue
		}

		_, executable := s.handlersMgr.CheckEventPath(zEvent.Namespace, zEvent.Type)
		if createSamples {
			s.handlersMgr.CheckSampleEvent(zEvent)
		}

		s.logger.Trace.Printf("Approved for execution: %v\n", zEvent)
		if executable {
			// Only put event on channel if event path exists as no handlers will be present //
			ch <- zEvent
		}
	}
}
