package libs

import (
    "encoding/json"
    "github.com/euforia/spinal-cord/reactor/revent"
    "github.com/euforia/spinal-cord/logging"
    zmq "github.com/pebbe/zmq3"
)

type Synapse struct {
    zsock *zmq.Socket
}

func NewSynapse(ztype zmq.Type, uri string) (*Synapse, error) {
    sock, _ := zmq.NewSocket(ztype)
    s := Synapse{sock}
    err := s.zsock.Connect(uri)
    if err != nil {
        return &s, err
    }
    return &s, nil
}

func (s *Synapse) Fire(event *revent.Event) (string, error) {

    defer s.zsock.Close()

    msg, err := event.JsonString()
    if err != nil {
        return "", err
    }
    s.zsock.Send(msg, 0)
    resp, err := s.zsock.Recv(0)
    if err != nil {
        return "", err
    }
    return resp, nil
}

func FireSynapse(logger *logging.Logger, config *Config) (string, error) {

    reqRepConfig := config.TypeConfig.(map[string]string)

    var payload map[string]interface{}
    err := json.Unmarshal([]byte(reqRepConfig["event_data"]), &payload)
    if err != nil {
        //logger.Error.Fatalf("Could not serialize data: %s\n", reqRepConfig["event_data"])
        return "", err
    }

    evt := revent.NewEvent(config.Namespace, reqRepConfig["event_type"], payload)

    synapse, err := NewSynapse(zmq.REQ, config.SpinalCord)
    if err != nil {
        //logger.Error.Fatal(err)
        return "", err
    }
    resp, err := synapse.Fire(evt)
    if err != nil {
        //logger.Error.Fatal(err)
        return "", nil
    }
    //fmt.Println(resp)
    return resp, nil
}
