package synapse

import (
	//"encoding/json"
	"github.com/euforia/spinal-cord/config"
	//"github.com/euforia/spinal-cord/logging"
	"fmt"
	revent "github.com/euforia/spinal-cord/revent/v2"
	zmq "github.com/pebbe/zmq3"
)

type ZMQSynapse struct {
	zsock *zmq.Socket
	ztype string
}

func NewZMQSynapse(cfg config.SynapseConfig) (*ZMQSynapse, error) {
	//func NewZMQSynapse(ztype string, uri string) (*ZMQSynapse, error) {
	var (
		//sock *zmq.Socket
		syn ZMQSynapse = ZMQSynapse{}
		err error
	)

	switch SynapseType(cfg.Type) {
	case SYN_TYPE_ZPUSH:
		syn.zsock, err = zmq.NewSocket(zmq.PUSH)
	case SYN_TYPE_ZREQ:
		syn.zsock, err = zmq.NewSocket(zmq.REQ)
	default:
		return &syn, fmt.Errorf("Type not supported: %s", cfg.Type)
	}
	syn.ztype = cfg.Type

	if err != nil {
		return &syn, err
	}

	if err = syn.zsock.Connect(cfg.SpinalUri); err != nil {
		return &syn, err
	}
	//syn.zsock = sock
	return &syn, nil
}

func (s *ZMQSynapse) Fire(event *revent.Event) error {
	/*
		if s.ztype == "REQ" {
			defer s.Close()
		}
	*/
	msg, err := event.JsonString()
	if err != nil {
		return err
	}
	d, err := s.zsock.Send(msg, 0)
	//resp, err := s.zsock.Recv(0)
	if err != nil {
		return err
	}
	if d < 0 {
		return fmt.Errorf("Failed to send: %d", d)
	}

	if s.ztype == "REQ" {
		_, err := s.zsock.Recv(0)
		if err != nil {
			return err
		}
	}
	//fmt.Println(msg)
	//return resp, nil
	return nil
}

func (s *ZMQSynapse) Close() {
	s.zsock.Close()
}

/*
func FireZMQSynapse(logger *logging.Logger, cfg *config.Config) (string, error) {

	reqRepConfig := cfg.TypeConfig.(map[string]string)

	var payload map[string]interface{}
	err := json.Unmarshal([]byte(reqRepConfig["event_data"]), &payload)
	if err != nil {
		//logger.Error.Fatalf("Could not serialize data: %s\n", reqRepConfig["event_data"])
		return "", err
	}

	evt := revent.NewEvent(cfg.Namespace, reqRepConfig["event_type"], payload)

	synapse, err := NewZMQSynapse(zmq.REQ, cfg.SpinalCord)
	if err != nil {
		//logger.Error.Fatal(err)
		return "", err
	}
	defer synapse.Close()

	resp, err := synapse.Fire(evt)
	if err != nil {
		//logger.Error.Fatal(err)
		return "", nil
	}
	//fmt.Println(resp)
	return resp, nil
}
*/
