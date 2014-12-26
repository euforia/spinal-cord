package v2

import (
	"os"
	"testing"
)

var testEventFile string = "/tmp/zmq.test.json"

func Test_NewEvent_WriteToFile(t *testing.T) {
	e := NewEvent(
		"testing",
		"zmq.test",
		map[string]interface{}{
			"name":        "NewEvent",
			"description": "Testing NewEvent.",
		})

	err := e.WriteToFile(testEventFile, 0777)
	if err != nil {
		t.Errorf("Failed to write event: %s\n", err)
		t.FailNow()
	}

	jstr, err := e.JsonString()
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	t.Logf("%s", jstr)
}

func Test_LoadEventFromFrile(t *testing.T) {
	_, err := LoadEventFromFile("filepath")
	if err == nil {
		t.Errorf("failed error check")
	}
	d, err := LoadEventFromFile(testEventFile)
	if err != nil {
		t.Errorf("%s\n", err)
		t.FailNow()
	}
	t.Logf("%v\n", d)

	os.Remove(testEventFile)
}
