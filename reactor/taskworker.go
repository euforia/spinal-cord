package reactor

import (
	"encoding/json"
	"fmt"
	"github.com/euforia/spinal-cord/logging"
	revent "github.com/euforia/spinal-cord/revent/v2"
	zmq "github.com/pebbe/zmq3"
)

type TaskWorker struct {
	/* send tasks received on socket to this channel (go func)*/
	taskChan chan Task
	/* collect results on this channel */
	Results chan map[string]interface{}
	/* recv tasks on this socket from TaskPusher */
	client      *zmq.Socket
	handlersMgr *HandlersManager
	logger      *logging.Logger
}

func NewTaskWorker(taskServerUri string, handlersDir string, logger *logging.Logger) (*TaskWorker, error) {
	tw := TaskWorker{
		taskChan:    make(chan Task),
		handlersMgr: NewHandlersManager(handlersDir, logger),
		logger:      logger,
		Results:     make(chan map[string]interface{}),
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
		tw.logger.Info.Printf("QUEUEING - Handler: %s...\n", recvTask.TaskHandler.Path)
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

//func (tw *TaskWorker) runHandler(runtask Task) {
func (tw *TaskWorker) runHandler(runtask Task, rsltChan chan map[string]interface{}) {
	result := runtask.Run(tw.handlersMgr.HandlersDir)
	val, ok := result["error"]
	if ok {
		tw.logger.Error.Printf("FAILED - Handler: %s; Message: %s", runtask.TaskHandler.Path, val)
	} else {
		tw.logger.Info.Printf("SUCCESS - Handler: %s\n", runtask.TaskHandler.Path)
		tw.logger.Debug.Printf("RESULT - Handler: %s => %v", runtask.TaskHandler.Path, result["data"])
	}
	rsltChan <- result
}

func (tw *TaskWorker) startProcessingWork(taskCh chan Task, rsltCh chan map[string]interface{}) {
	for {
		task := <-taskCh
		err := tw.runPreExecutionChecks(task)
		if err != nil {
			tw.logger.Error.Println(err)
			continue
		}
		tw.logger.Info.Printf("EXECUTING - Handler: %s\n", task.TaskHandler.Path)
		tw.logger.Debug.Printf("EXECUTING - Handler: %s; Payload: %s\n",
			task.TaskHandler.Path, task.Payload)

		/*
		 * TODO: ?? check spawn count or wait ??
		 * May be just set GOMAXPROCS
		 */
		go tw.runHandler(task, rsltCh)
		//go tw.runHandler(task)
	}
}

func (tw *TaskWorker) Start() {
	go tw.startReceivingWork(tw.taskChan)
	go tw.startProcessingWork(tw.taskChan, tw.Results)
}
