package libs

import(
    "fmt"
    "testing"
    zmq "github.com/pebbe/zmq3"
    "github.com/euforia/spinal-cord/reactor/revent"
)

var testSpinalCordUri = "tcp://localhost:55055"
var testEvent *revent.Event = revent.NewEvent(
                                    "testing",
                                    "zmq.test",
                                    map[string]interface{}{
                                        "name": "Synapse",
                                        "description": "Testing synapse fire.",
                                    })

func Test_NewSynapse(t *testing.T) {
    _, err := NewSynapse(zmq.REQ, testSpinalCordUri)
    if err != nil {
        t.Error(fmt.Sprintf("Failed to create synapse: %s %s", testSpinalCordUri, err))
    }
}
func Test_Synapse_Fire(t *testing.T) {
    s, err := NewSynapse(zmq.REQ, testSpinalCordUri)
    if err != nil {
        t.Error(fmt.Sprintf("Failed to create synapse: %s %s", testSpinalCordUri, err))
    }

    _, err = s.Fire(testEvent)
    if err != nil {
        t.Errorf("Failed to fire synapse: %s %s", testSpinalCordUri, err)
    }
}
