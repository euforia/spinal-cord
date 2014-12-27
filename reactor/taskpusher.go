package reactor

import (
	"encoding/json"
	"errors"
	"fmt"
	revent "github.com/euforia/spinal-cord/revent/v2"
	zmq "github.com/pebbe/zmq3"
)

type TaskPusher struct {
	zsock *zmq.Socket
}

func NewTaskPusher(listenAddr string) (*TaskPusher, error) {
	var (
		t   = TaskPusher{}
		err error
	)
	t.zsock, err = zmq.NewSocket(zmq.PUSH)
	if err != nil {
		return &t, err
	}
	if err = t.zsock.Bind(listenAddr); err != nil {
		return &t, err
	}
	return &t, nil
}

func (t *TaskPusher) EncodeEvent(evt revent.Event) (string, error) {
	b, err := json.Marshal(&evt)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t *TaskPusher) PushTask(evt revent.Event, evtTaskHandler EventHandler) error {
	//func (s *SubReactor) assembleTask(evtBytes []byte, evtHandler EventHandler) (string, error) {
	hdlr, err := evtTaskHandler.Handler()
	if err != nil {
		return err
	}
	encEvt, err := t.EncodeEvent(evt)
	if err != nil {
		return err
	}
	newtask := Task{encEvt, hdlr}
	taskBytes, err := newtask.Serialize()
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't serialize task: %v, reason: %v", newtask, err))
	}

	if _, err = t.zsock.SendBytes(taskBytes, 0); err != nil {
		return err
	}
	return nil
}
