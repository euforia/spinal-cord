package handler

import(
    "errors"
    "os"
    "fmt"
    "github.com/euforia/spinal-cord/logging"
    "github.com/euforia/spinal-cord/reactor/revent"
    "io/ioutil"
    "crypto/sha1"
)


type EventHandlersDetails struct {
    Namespace string        `json:"namespace"`
    Type string             `json:"event_type"`
    Sample revent.Event     `json:"sample"`
    Handlers []EventHandler `json:"handlers"`
}

type HandlersManager struct {
    HandlersDir string
    logger *logging.Logger
}

func NewHandlersManager(dirPath string, logger *logging.Logger) *HandlersManager {
    return &HandlersManager{dirPath, logger}
}

func (hm *HandlersManager) GetHandlers(ns string, etype string) []EventHandler {
    handlers := make([]EventHandler, 0)

    eventPath := fmt.Sprintf("%s/%s/%s", hm.HandlersDir, ns, etype)
    handlerScripts, _ := ioutil.ReadDir(eventPath)

    sampleEvtFile := fmt.Sprintf("%s.json", etype)
    for _, h := range handlerScripts {
        if h.Name() == sampleEvtFile {
            continue
        }
        handlers = append(handlers, EventHandler{
                                        fmt.Sprintf("%s/%s/%s", ns, etype, h.Name()),
                                        fmt.Sprintf("%s/%s", eventPath, h.Name()),
                                        h.Name()})
    }

    return handlers
}

func (hm *HandlersManager) fileExists(path string) bool {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    return false
}

func (hm *HandlersManager) EventDetails(ns string, etype string) (EventHandlersDetails, error) {

    var eDetails EventHandlersDetails
    fullpath, exists := hm.EventPath(ns, etype)
    if exists {
        samplepath := fmt.Sprintf("%s/%s.json", fullpath, etype)
        sampleEvent, err := revent.LoadEvent(samplepath)
        if err != nil {
            hm.logger.Error.Println(err)
            return eDetails, err
        }

        handlers := hm.GetHandlers(ns, etype)
        eDetails = EventHandlersDetails{ns, etype, sampleEvent, handlers}
        return eDetails, nil
    }
    return eDetails, fmt.Errorf(fmt.Sprintf("Event path not found: %s/%s", ns, etype))
}

func (hm *HandlersManager) EventPath(ns string, etype string) (string, bool) {
    abspath := fmt.Sprintf("%s/%s/%s", hm.HandlersDir, ns, etype)
    return abspath, hm.fileExists(abspath)
}

func (hm *HandlersManager) PathExists(iPath string) bool {
    abspath := fmt.Sprintf("%s/%s", hm.HandlersDir, iPath)
    return hm.fileExists(abspath)
}


func (hm *HandlersManager) CheckSampleEvent(evt revent.Event) {
    filepath := fmt.Sprintf("%s/%s/%s/%s.json", hm.HandlersDir, evt.Namespace, evt.Type, evt.Type)
    if hm.fileExists(filepath) {
        return
    }
    err := evt.WriteToFile(filepath, 0755)
    if err != nil {
        hm.logger.Error.Println("Could not write sample event:", err)
    }
    hm.logger.Warning.Printf("Wrote sample event: %s/%s/%s.json\n", evt.Namespace, evt.Type, evt.Type)
}

/*
    Return:
        created bool
        executable bool
*/
func (hm *HandlersManager) CheckEventPath(ns string, etype string) (bool, bool) {

    eventPath, exists := hm.EventPath(ns, etype)
    if !exists {
        hm.logger.Warning.Printf("Creating event path: %s/%s\n", ns, etype)
        err := os.MkdirAll(eventPath, 0777)
        if err != nil {
            hm.logger.Error.Println(err)
            return false, false
        }
        return true, false
    }
    return false, true
}

func (hm *HandlersManager) GetHandler(handlerPath string) (*Handler, error) {
    absPath := fmt.Sprintf("%s/%s", hm.HandlersDir, handlerPath)
    if hm.fileExists(absPath) {
        return GetHandlerFromFile(absPath)
    }
    return nil, fmt.Errorf("handler not found: %s", handlerPath)
}

func (hm *HandlersManager) CheckHandler(handler Handler) error {

    absPath := fmt.Sprintf("%s/%s", hm.HandlersDir, handler.Path)

    if hm.fileExists(absPath) {
        dbytes, _ := ioutil.ReadFile(absPath)
        dbytesSha1 := fmt.Sprintf("%x", sha1.Sum(dbytes))
        if dbytesSha1 == handler.Sha1String() {
            hm.logger.Trace.Println("Handler up-to-date:", handler.Path)
            return nil
        }
    }

    calcSha1 := fmt.Sprintf("%x", sha1.Sum(handler.Data))
    if calcSha1 != handler.Sha1String() {
        return errors.New(fmt.Sprintf("Not writing handler - SHA1 mismatch: %s", handler.Path))
    }

    // write handler //
    err := handler.WriteHandlerFile(hm.HandlersDir, 0777)
    if err != nil {
        return errors.New(fmt.Sprintf("Failed to write handler: %s; reason: %v\n", handler.Path, err))
    }
    hm.logger.Warning.Printf("Updated - handler: %s; sha1: %x\n", handler.Path, handler.Sha1)
    return nil
}

func (hm *HandlersManager) Namespaces() ([]string, error) {
    dirnames := make([]string, 0)
    listing, err:= ioutil.ReadDir(hm.HandlersDir)
    if err != nil {
        return dirnames, err
    }
    for _, f := range listing {
        if f.IsDir() {
            dirnames = append(dirnames, f.Name())
        }
    }
    return dirnames, nil
}

func (hm *HandlersManager) EventTypes(namespace string) ([]string, error) {
    eventTypes := make([]string, 0)

    listing, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", hm.HandlersDir, namespace))
    if err != nil {
        return eventTypes, err
    }
    for _, e := range listing {
        eventTypes = append(eventTypes, e.Name())
    }
    return eventTypes, nil
}


