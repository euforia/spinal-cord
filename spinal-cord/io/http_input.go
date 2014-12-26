package io

import (
	"encoding/json"
	"fmt"
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	revent "github.com/euforia/spinal-cord/revent/v2"
	//"github.com/euforia/spinal-cord/synapse"
	"io/ioutil"
	"net/http"
)

const ACL_DEFAULT_ORIGIN string = "*"

type HttpSpinalInput struct {
	endpoint   string
	pubChannel chan revent.Event
	logger     *logging.Logger
}

func NewHttpSpinalInput(cfg config.IOConfig, logger *logging.Logger) (*HttpSpinalInput, error) {
	var (
		h  HttpSpinalInput = HttpSpinalInput{logger: logger}
		ok bool
	)

	if h.endpoint, ok = cfg.Config["endpoint"].(string); !ok {
		return &h, fmt.Errorf(fmt.Sprintf("invalid endpoint type :%s\n", cfg.Config["endpoint"]))
	}
	return &h, nil
}

func (h *HttpSpinalInput) checkRequest(r *http.Request) (*revent.Event, error) {
	//var annoReq annotations.EventAnnotation
	var evt revent.Event

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &evt, err
	}
	//h.logger.Trace.Printf("Checking event: %s\n", body)

	err = json.Unmarshal(body, &evt)
	if err != nil {
		return &evt, err
	}
	err = evt.Validate()
	if err != nil {
		return &evt, err
	}
	h.logger.Trace.Printf("Event validated: %s\n", evt)
	return &evt, nil
}

func (h *HttpSpinalInput) endpointPubHandler(w http.ResponseWriter, r *http.Request) {
	var (
		evt  *revent.Event
		resp interface{}
		code int
		err  error
	)
	switch r.Method {
	case "POST":
		evt, err = h.checkRequest(r)
		if err != nil {
			resp = fmt.Sprintf(`{"error": "%s"}`, err)
			code = 400
			break
		}
		resp = evt
		code = 200
		break
	default:
		resp = fmt.Sprintf(`{"error":"Method not supported: %s"}`, r.Method)
		code = 405
		break
	}
	h.writeJsonResponse(w, r, resp, code)

	if code == 200 {
		h.pubChannel <- *evt
	}
}

func (h *HttpSpinalInput) writeJsonResponse(w http.ResponseWriter, r *http.Request, data interface{}, respCode int) {
	var b []byte

	switch data.(type) {
	case []byte:
		b, _ = data.([]byte)
		break
	case string:
		ab, _ := data.(string)
		b = []byte(ab)
		break
	default:
		b, _ = json.Marshal(&data)
		break
	}

	w.Header().Set("Access-Control-Allow-Origin", ACL_DEFAULT_ORIGIN)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(respCode)
	w.Write(b)
	h.logger.Info.Printf("%s %d %s\n", r.Method, respCode, r.URL.RequestURI())
}

func (h *HttpSpinalInput) Start(ch chan revent.Event) {
	/*
	 * Not required as this runs as part of the main http server
	 */
	h.pubChannel = ch
	http.HandleFunc(h.endpoint, h.endpointPubHandler)
}

func (h *HttpSpinalInput) Stop() {
	/*
	 * Not required as this runs as part of the main http server
	 */
}
