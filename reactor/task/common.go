package task

import(
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "github.com/euforia/spinal-cord/reactor/handler"
)

const EVENT_ENV_NAME string = "REVENT"

type Task struct {
    Payload string `json:"payload"`
    TaskHandler handler.Handler `json:"handler"`
}

func (t *Task) Serialize() ([]byte, error) {
    return json.Marshal(t)
}

func (t *Task) CheckHandler(handlersPath string) error {

    fullPath := fmt.Sprintf("%s/%s", handlersPath, t.TaskHandler.Path)
    fileinfo, _ := os.Stat(fullPath)

    // check exec perms //
    if fileinfo.Mode() & 0111 == 0 {
        return errors.New(fmt.Sprintf("Handler not executable: '%s'", t.TaskHandler.Path))
    }
    return nil
}

func (t *Task) Run(handlersPath string) map[string] interface{} {

    fullPath := fmt.Sprintf("%s/%s", handlersPath, t.TaskHandler.Path)
    err := t.CheckHandler(handlersPath)
    if err != nil {
        return map[string]interface{}{"error": err}
    }

    os.Setenv(EVENT_ENV_NAME, t.Payload)
    output, err := exec.Command(fullPath).CombinedOutput()
    if err != nil {
        //return map[string]interface{}{"error":err}
        return map[string]interface{}{"error":output}
    }
    return map[string]interface{}{"data":string(output)}
}