package revent

import(
    "testing"
    "os"
)

var testEventFile string = "/tmp/test_zmq.test.json"

func Test_NewEvent_WriteToFile(t *testing.T) {
    e := NewEvent(
            "testing",
            "zmq.test",
            map[string]interface{}{
                "name": "NewEvent",
                "description": "Testing NewEvent.",
            })

    err := e.WriteToFile(testEventFile, 0777)
    if err != nil {
        t.Errorf("Failed to write event: %s\n", err)
    } else {
        os.Remove(testEventFile)
    }
}