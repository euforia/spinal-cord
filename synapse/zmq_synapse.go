package synapse

import (
	"encoding/json"
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
	var (
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

	if err = syn.zsock.Connect(cfg.URI); err != nil {
		return &syn, err
	}
	return &syn, nil
}

func (s *ZMQSynapse) Receive() (*revent.Event, error) {
	var e *revent.Event
	return e, fmt.Errorf("Fire only synapse")
}

func (s *ZMQSynapse) Fire(event *revent.Event) error {

	msg, err := event.JsonString()
	if err != nil {
		return err
	}
	d, err := s.zsock.Send(msg, 0)
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
	return nil
}

func (s *ZMQSynapse) Close() {
	s.zsock.Close()
}

type ZMQSubscriberSynapse struct {
	ZMQSynapse
}

func NewZMQSubscriberSynapse(cfg config.SynapseConfig) (*ZMQSubscriberSynapse, error) {
	var (
		zsubSyn = ZMQSubscriberSynapse{}
		err     error
	)
	zsubSyn.ztype = "SUB"
	zsubSyn.zsock, err = zmq.NewSocket(zmq.SUB)
	if err != nil {
		return &zsubSyn, err
	}
	if err = zsubSyn.zsock.Connect(cfg.URI); err != nil {
		return &zsubSyn, err
	}

	mcfg, ok := cfg.TypeConfig.(map[string]interface{})
	if !ok {
		return &zsubSyn, fmt.Errorf("invalid config type: %s", cfg.TypeConfig)
	}

	tcfg, err := zsubSyn.getTypeConfig(mcfg)
	if err != nil {
		return &zsubSyn, err
	}

	zsubSyn.setSubscriptions(tcfg.Subscriptions)

	return &zsubSyn, nil
}

func (z *ZMQSubscriberSynapse) getTypeConfig(ifc map[string]interface{}) (config.RecvSynapseConfig, error) {
	var (
		r  = config.RecvSynapseConfig{}
		ok bool
	)

	r.Type, ok = ifc["type"].(string)
	if !ok {
		return r, fmt.Errorf("invalid config type: %s", ifc["type"])
	}

	subs, ok := ifc["subscriptions"].([]interface{})
	if !ok {
		return r, fmt.Errorf("invalid config type: %s", ifc["subscriptions"])
	}
	r.Subscriptions = make([]string, len(subs))
	for i, s := range subs {
		str, ok := s.(string)
		if !ok {
			return r, fmt.Errorf("invalid config type: %s", s)
		}
		r.Subscriptions[i] = str
	}
	return r, nil
}

func (z *ZMQSubscriberSynapse) setSubscriptions(subs []string) {
	if len(subs) > 0 {
		for _, s := range subs {
			z.zsock.SetSubscribe(s)
		}
	} else {
		z.zsock.SetSubscribe("")
	}
}

func (z *ZMQSubscriberSynapse) Receive() (*revent.Event, error) {
	var (
		evt revent.Event
		err error
		b   []byte
	)

	if b, err = z.zsock.RecvBytes(0); err == nil {
		if err = json.Unmarshal(b, &evt); err == nil {
			return &evt, nil
		}
		return &evt, err
	}
	return &evt, err
}

func (z *ZMQSubscriberSynapse) Fire(evt *revent.Event) error {
	return nil
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
