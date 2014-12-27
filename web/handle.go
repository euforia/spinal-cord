package web

import (
	"fmt"
	"github.com/euforia/spinal-cord/reactor"
	"net/http"
)

type BaseSpinalCordHandle struct {
	DefaultEndpointHandler
	mgr *reactor.HandlersManager
}

func NewBaseSpinalCordHandle(mgr *reactor.HandlersManager) BaseSpinalCordHandle {
	return BaseSpinalCordHandle{DefaultEndpointHandler{}, mgr}
}

type NamespaceHandle struct {
	BaseSpinalCordHandle
}

func NewNamespaceHandle(mgr *reactor.HandlersManager) *NamespaceHandle {
	return &NamespaceHandle{NewBaseSpinalCordHandle(mgr)}
}
func (n *NamespaceHandle) GET(r *http.Request, args ...string) (interface{}, int) {
	nss, err := n.mgr.Namespaces()
	if err != nil {
		return map[string]string{"error": fmt.Sprintf("%s", err)}, 500
	}
	return nss, 200
}

type EventTypeHandle struct {
	BaseSpinalCordHandle
}

func NewEventTypeHandle(mgr *reactor.HandlersManager) *EventTypeHandle {
	return &EventTypeHandle{NewBaseSpinalCordHandle(mgr)}
}
func (n *EventTypeHandle) GET(r *http.Request, args ...string) (interface{}, int) {
	rslt, err := n.mgr.EventTypes(args[0])
	if err != nil {
		return map[string]string{"error": "Event type not found!"}, 404
	}
	return rslt, 200
}

type EventTypeHandlersHandle struct {
	BaseSpinalCordHandle
}

func NewEventTypeHandlersHandle(mgr *reactor.HandlersManager) *EventTypeHandlersHandle {
	return &EventTypeHandlersHandle{NewBaseSpinalCordHandle(mgr)}
}
func (n *EventTypeHandlersHandle) GET(r *http.Request, args ...string) (interface{}, int) {
	rslt, err := n.mgr.EventDetails(args[0], args[1])
	if err != nil {
		return map[string]string{"error": fmt.Sprintf("Event type not found: %s:%s!",
			args[0], args[1])}, 404
	}
	return rslt, 200
}

type EventHandlerHandle struct {
	BaseSpinalCordHandle
}

func NewEventHandlerHandle(mgr *reactor.HandlersManager) *EventHandlerHandle {
	return &EventHandlerHandle{NewBaseSpinalCordHandle(mgr)}
}

func (n *EventHandlerHandle) checkWriteRequestData(r *http.Request) map[string]interface{} {
	data, err := n.JsonBody(r)
	if err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("%s", err)}
	}
	_, ok := data["content"]
	if !ok {
		return map[string]interface{}{"error": "Required key: 'content'"}
	}
	fmt.Println(data)
	return data
}

func (n *EventHandlerHandle) writeHandlerToFile(iRequestPath string, data string, overwrite bool) map[string]string {
	ihandler, err := reactor.NewHandler(iRequestPath, []byte(data))
	if err != nil {
		return map[string]string{"error": fmt.Sprintf("%s", err)}
	}
	if !overwrite {
		if exists := n.mgr.PathExists(iRequestPath); exists {
			return map[string]string{"error": fmt.Sprintf("Path exists: %s", iRequestPath)}
		}
	}
	err = ihandler.WriteHandlerFile(n.mgr.HandlersDir, 0777)
	if err != nil {
		return map[string]string{"error": fmt.Sprintf("%s", err)}
	}
	return map[string]string{"status": "success"}
}
func (n *EventHandlerHandle) GET(r *http.Request, args ...string) (interface{}, int) {
	ehdlr, err := n.mgr.GetHandler(r.URL.Path)
	if err != nil {
		return map[string]string{"error": fmt.Sprintf("%s", err)}, 500
	}
	return map[string]string{"data": string(ehdlr.Data),
		"sha1": ehdlr.Sha1String(),
		"path": r.URL.Path}, 200
}

func (n *EventHandlerHandle) POST(r *http.Request, args ...string) (interface{}, int) {
	mReq := n.checkWriteRequestData(r)
	_, ok := mReq["error"]
	if ok {
		return mReq, 400
	}
	status := n.writeHandlerToFile(r.URL.Path, fmt.Sprintf("%s", mReq["content"]), false)
	_, ok = status["error"]
	if ok {
		return status, 400
	}
	return status, 200
}

func (n *EventHandlerHandle) PUT(r *http.Request, args ...string) (interface{}, int) {
	mReq := n.checkWriteRequestData(r)
	_, ok := mReq["error"]
	if ok {
		return mReq, 400
	}
	status := n.writeHandlerToFile(r.URL.Path, fmt.Sprintf("%s", mReq["content"]), true)
	_, ok = status["error"]
	if ok {
		return status, 400
	}
	return status, 200
}

func (n *EventHandlerHandle) DELETE(r *http.Request, args ...string) (interface{}, int) {
	if !n.mgr.PathExists(r.URL.Path) {
		return map[string]string{"error": "not found"}, 404
	}
	ehdlr, err := n.mgr.GetHandler(r.URL.Path[1:])
	if err != nil {
		return map[string]string{"error": fmt.Sprintf("%s", err)}, 500
	}
	err = ehdlr.Remove()
	if err != nil {
		return map[string]string{"error": fmt.Sprintf("%s", err)}, 500
	}
	return map[string]string{"status": "success"}, 200
}
