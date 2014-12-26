package v2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type Event struct {
	Namespace string                 `json:"namespace"`
	Type      string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp float64                `json:"timestamp"`
}

func NewEvent(ns, etype string, data map[string]interface{}) *Event {
	return &Event{
		Namespace: ns,
		Type:      etype,
		Data:      data,
		Timestamp: float64(time.Now().UnixNano()) / 1000000000}
}

func (e *Event) Validate() error {
	if e.Namespace == "" {
		return fmt.Errorf("Namespace required")
	}
	if e.Type == "" {
		return fmt.Errorf("Type required")
	}
	if e.Timestamp <= 0 {
		e.Timestamp = float64(time.Now().UnixNano()) / 1000000000
	}
	return nil
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

func LoadEventFromFile(filepath string) (Event, error) {
	var evt Event
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return evt, err
	}
	if err = json.Unmarshal(data, &evt); err != nil {
		return evt, err
	}
	return evt, nil
}
