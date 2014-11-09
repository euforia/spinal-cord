package reactor


import(
    "fmt"
    "encoding/json"
    "errors"
    "github.com/euforia/spinal-cord/reactor/task"
    "github.com/euforia/spinal-cord/logging"
    "github.com/euforia/spinal-cord/reactor/handler"
    "github.com/euforia/spinal-cord/reactor/revent"
    zmq "github.com/pebbe/zmq3"
)

type SubReactor struct {

    Aggregator *zmq.Socket
    TaskServer * zmq.Socket

    handlersMgr *handler.HandlersManager

    logger *logging.Logger
}

func NewSubReactor(aggrUri string, listenAddr string, handlersDir string, logger *logging.Logger) *SubReactor {

    aggr, _ := zmq.NewSocket(zmq.SUB)
    aggr.Connect(aggrUri)
    aggr.SetSubscribe("")
    logger.Warning.Printf("Connected to: %s\n", aggrUri)

    server, _ := zmq.NewSocket(zmq.PUSH)
    server.Bind(listenAddr)
    logger.Warning.Printf("Task server started: %s\n", listenAddr)

    return &SubReactor{aggr, server, handler.NewHandlersManager(handlersDir, logger), logger}
}

func (s *SubReactor) Start(createSamples bool) {
    commChan := make(chan revent.Event)

    go s.startConsumingFromAggregator(commChan, createSamples)

    s.startPushingToWorkers(commChan)
}

func (s *SubReactor) decodeMessage(msg []string) (revent.Event, error) {
    var event revent.Event
    switch(len(msg)) {
        case 1:
            err := json.Unmarshal([]byte(msg[0]), &event)
            if err != nil { return event, err }
            break
        case 2:
            err := json.Unmarshal([]byte(msg[1]), &event)
            if err != nil { return event, err }
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

        event := <- ch
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
        s.logger.Trace.Printf("Sending event to worker: %v\n", zEvent)
        if executable {
            // Only put event on channel if event path exists as no handlers will be present //
            ch <- zEvent
        }
    }
}
