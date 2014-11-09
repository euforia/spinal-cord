package task

import(
    "encoding/json"
    "github.com/euforia/spinal-cord/logging"
    "github.com/euforia/spinal-cord/reactor/revent"
    "github.com/euforia/spinal-cord/reactor/handler"
    zmq "github.com/pebbe/zmq3"
)

type TaskWorker struct {
    client *zmq.Socket
    handlersMgr *handler.HandlersManager
    logger *logging.Logger
}

func NewTaskWorker(taskServerUri string, handlersDir string, logger *logging.Logger) *TaskWorker {
    client, _ := zmq.NewSocket(zmq.PULL)
    hdlMgr := handler.NewHandlersManager(handlersDir, logger)

    tw := TaskWorker{client, hdlMgr, logger}

    err := tw.client.Connect(taskServerUri)
    if err != nil {
        logger.Error.Fatalf("Connection failed to task server: %s; reason: %v\n", taskServerUri, err)
    }
    logger.Warning.Printf("Connected to task server: %s\n", taskServerUri)

    return &tw
}

func (tw *TaskWorker) getTask() (Task, error) {
    var recvdTask Task

    msg, err := tw.client.Recv(0)
    if err != nil {
        return recvdTask, err
    }

    err = json.Unmarshal([]byte(msg), &recvdTask)
    if err != nil {
        return recvdTask, err
    }
    return recvdTask, nil
}

func (tw *TaskWorker) startReceivingWork(ch chan Task) {
    for {
        recvTask, err := tw.getTask()
        if err != nil {
            tw.logger.Warning.Println(err)
            continue
        }
        tw.logger.Info.Printf("Received task: %v\n", recvTask)
        tw.logger.Info.Printf("Queueing - handler: %s...\n", recvTask.TaskHandler.Path)
        ch <- recvTask
    }
}


func (tw *TaskWorker) runPreExecutionChecks(ctask Task) error {

    var evt revent.Event
    err := json.Unmarshal([]byte(ctask.Payload), &evt)
    if err != nil {
        return err
    }
    // checks //
    tw.handlersMgr.CheckEventPath(evt.Namespace, evt.Type)
    // copy event handler to worker if sha1 mismatch or missing //
    err = tw.handlersMgr.CheckHandler(ctask.TaskHandler)
    if err != nil {
        return err
    }
    return nil
}

func (tw *TaskWorker) runHandler(runtask Task) {

    result := runtask.Run(tw.handlersMgr.HandlersDir)
    val, ok := result["error"]
    if ok {
        tw.logger.Error.Printf("handler: %s message: %s", runtask.TaskHandler.Path, val)
    } else {
        tw.logger.Info.Printf("Execution complete - handler: %s\n", runtask.TaskHandler.Path)
        tw.logger.Debug.Printf("Result - handler: %s; => %v", runtask.TaskHandler.Path, result["data"])
    }
}

func (tw *TaskWorker) startProcessingWork(ch chan Task) {
    for {
        task := <- ch
        err := tw.runPreExecutionChecks(task)
        if err != nil {
            tw.logger.Error.Println(err)
            continue
        }
        tw.logger.Debug.Printf("Executing Handler: %s; with Payload: %s\n",
                                        task.TaskHandler.Path, task.Payload)

        /* TODO: ?? check spawn count or wait ?? */
        go tw.runHandler(task)
    }
}

func (tw *TaskWorker) Start() {
    commChan := make(chan Task)

    go tw.startReceivingWork(commChan)
    tw.startProcessingWork(commChan)
}