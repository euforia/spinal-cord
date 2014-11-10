package task

import(
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "time"
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
        //t.WriteLog(handlersPath, []byte(fmt.Sprintf("%s", err)), false)
        return map[string]interface{}{"error": err}
    }

    os.Setenv(EVENT_ENV_NAME, t.Payload)
    output, err := exec.Command(fullPath).CombinedOutput()
    if err != nil {
        //t.WriteLog(handlersPath, []byte(output), false)
        return map[string]interface{}{"error":output}
    }
    //t.WriteLog(handlersPath, []byte(output), true)
    return map[string]interface{}{"data":string(output)}
}

func (t *Task) WriteLog(handlersPath string, data []byte, success bool) error {

    var logfile string
    if success {
        logfile = fmt.Sprintf("%s/%s.%d.log", handlersPath, t.TaskHandler.Path, time.Now().Unix())
    } else {
        logfile = fmt.Sprintf("%s/%s.%d.error.log", handlersPath, t.TaskHandler.Path, time.Now().Unix())
    }
    return ioutil.WriteFile(logfile, data, 0755)
}