package reactor

import (
	//"encoding/json"
	//"errors"
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	//"github.com/euforia/spinal-cord/reactor/handler"
	//"github.com/euforia/spinal-cord/reactor/task"
	revent "github.com/euforia/spinal-cord/revent/v2"
	"github.com/euforia/spinal-cord/synapse"
	//zmq "github.com/pebbe/zmq3"
)

type SubReactor struct {
	recvSynapse synapse.ISynapse

	//TaskServer *zmq.Socket
	taskPusher *TaskPusher

	handlersMgr *HandlersManager

	logger *logging.Logger
}

func NewSubReactor(cfg *config.SpinalCordConfig, logger *logging.Logger) (*SubReactor, error) {
	/* listen for published messages */
	var (
		sreactor   SubReactor = SubReactor{}
		err        error
		listenAddr = fmt.Sprintf("tcp://*:%d", cfg.Reactor.Port)
	)

	if sreactor.recvSynapse, err = synapse.LoadSynapse(cfg.Reactor.Synapse); err != nil {
		return &sreactor, err
	}
	sreactor.handlersMgr = NewHandlersManager(cfg.Core.HandlersDir, logger)
	sreactor.logger = logger

	/* task worker server */
	if sreactor.taskPusher, err = NewTaskPusher(listenAddr); err != nil {
		return &sreactor, err
	}

	logger.Warning.Printf("Task server started: PUSH %s\n", listenAddr)

	return &sreactor, nil
}

func (s *SubReactor) Start(createSamples bool) {
	s.logger.Info.Printf("Create samples: %s", createSamples)

	commChan := make(chan revent.Event)

	go s.startConsumingFromAggregator(commChan, createSamples)

	s.startPushingToWorkers(commChan)
}

func (s *SubReactor) startPushingToWorkers(ch chan revent.Event) {
	for {
		s.logger.Trace.Println("Waiting for event...")

		event := <-ch

		handlers := s.handlersMgr.GetHandlers(event.Namespace, event.Type)

		s.logger.Info.Printf("Namespace => %s; Event => %s; Handlers => %d\n",
			event.Namespace, event.Type, len(handlers))

		for _, evtHandler := range handlers {
			if err := s.taskPusher.PushTask(event, evtHandler); err != nil {
				s.logger.Error.Printf("Failed to push task: %s\n", err)
			}
		}
	}
}

func (s *SubReactor) startConsumingFromAggregator(ch chan revent.Event, createSamples bool) {
	for {
		// Get event from zmq PUB queue //
		evt, err := s.recvSynapse.Receive()
		if err != nil {
			s.logger.Error.Println("Receive:", err)
			continue
		}

		_, executable := s.handlersMgr.CheckEventPath(evt.Namespace, evt.Type)
		if createSamples {
			s.handlersMgr.CheckSampleEvent(*evt)
		}

		s.logger.Trace.Printf("Approved for execution: %#v\n", evt)
		if executable {
			// Only put event on channel if event path exists as no handlers will be present //
			ch <- *evt
		}
	}
}
