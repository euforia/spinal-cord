package revent

import(
    "encoding/json"
    "io/ioutil"
    "os"
    "time"
)

type Event struct {
    Namespace string               `json:"namespace"`
    Type string                    `json:"event_type"`
    Payload map[string]interface{} `json:"payload"`
    Timestamp interface{}          `json:"timestamp"`
}

func NewEvent(ns, etype string, payload map[string]interface{}) *Event {
    return &Event{
        Namespace: ns,
        Type: etype,
        Payload: payload,
        Timestamp: float64(time.Now().UnixNano())/1000000000,}
}

func (e *Event) WriteToFile(filepath string, perms os.FileMode) error {
    data, err := json.MarshalIndent(&e, "", "  ")
    if err != nil {
        return err
    }
    return ioutil.WriteFile(filepath, data, perms)
}

func (e *Event) JsonString() (string, error) {
    bytes, err := json.Marshal(e)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}

func LoadEvent(filepath string) (Event, error) {
    var evt Event
    data, err := ioutil.ReadFile(filepath)
    if err != nil {
        return evt, err
    }
    if err = json.Unmarshal(data, &evt); err != nil {
        return evt, nil
    }
    return evt, err
}