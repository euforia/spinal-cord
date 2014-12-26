package task

import (
	"encoding/json"
	"fmt"
	"github.com/euforia/spinal-cord/logging"
	"github.com/euforia/spinal-cord/reactor/handler"
	revent "github.com/euforia/spinal-cord/revent/v2"
	zmq "github.com/pebbe/zmq3"
)

type TaskWorker struct {
	taskChan    chan Task
	client      *zmq.Socket
	handlersMgr *handler.HandlersManager
	logger      *logging.Logger
}

func NewTaskWorker(taskServerUri string, handlersDir string, logger *logging.Logger) (*TaskWorker, error) {
	tw := TaskWorker{
		taskChan:    make(chan Task),
		handlersMgr: handler.NewHandlersManager(handlersDir, logger),
		logger:      logger,
	}
	/* will receive tasks on this connection */
	tw.client, _ = zmq.NewSocket(zmq.PULL)

	err := tw.client.Connect(taskServerUri)
	if err != nil {
		return &tw, fmt.Errorf("Connection failed to task server: %s; reason: %v\n", taskServerUri, err)
	}
	logger.Warning.Printf("Connected to task server: %s\n", taskServerUri)

	return &tw, nil
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

/*
	This is run as a go function so channel must be passed in and not use
	the struct once.
*/
func (tw *TaskWorker) startReceivingWork(ch chan Task) {
	for {
		recvTask, err := tw.getTask()
		if err != nil {
			tw.logger.Warning.Println(err)
			continue
		}
		tw.logger.Debug.Printf("Received task: %v\n", recvTask)
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
		tw.logger.Error.Printf("FAILED - Handler: %s; Message: %s", runtask.TaskHandler.Path, val)
	} else {
		//tw.logger.Info.Printf("Execution complete - handler: %s\n", runtask.TaskHandler.Path)
		tw.logger.Info.Printf("SUCCESS - Handler: %s => %v", runtask.TaskHandler.Path, result["data"])
	}
}

func (tw *TaskWorker) startProcessingWork() {
	for {
		task := <-tw.taskChan
		err := tw.runPreExecutionChecks(task)
		if err != nil {
			tw.logger.Error.Println(err)
			continue
		}
		tw.logger.Debug.Printf("EXECUTING - Handler: %s; Payload: %s\n",
			task.TaskHandler.Path, task.Payload)

		/* TODO: ?? check spawn count or wait ?? */
		go tw.runHandler(task)
	}
}

func (tw *TaskWorker) Start() {
	go tw.startReceivingWork(tw.taskChan)
	tw.startProcessingWork()
}
