package revent

import(
    "encoding/json"
    "io/ioutil"
    "os"
)

type Event struct {
    Namespace string               `json:"namespace"`
    Type string                    `json:"event_type"`
    Payload map[string]interface{} `json:"payload"`
    Timestamp interface{}          `json:"timestamp"`
}

func (e *Event) WriteToFile(filepath string, perms os.FileMode) error {
    data, err := json.MarshalIndent(&e, "", "  ")
    if err != nil {
        return err
    }
    return ioutil.WriteFile(filepath, data, perms)
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