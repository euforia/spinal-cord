package web

import(
    "encoding/json"
    "fmt"
    "strings"
    "net/http"
    "github.com/euforia/spinal-cord/logging"
)

type RESTRouter struct {
    Prefix string
    defaultACLs string
    handlerMap []RESTEndpointHandler
    logger *logging.Logger
}

func NewRESTRouter(prefix string, defaultACLs string, logger *logging.Logger) *RESTRouter {
    return &RESTRouter{prefix, defaultACLs, make([]RESTEndpointHandler, 0), logger}
}

func (s *RESTRouter) writeHttpResponse(w http.ResponseWriter, headers map[string]string, data []byte, respCode int) {
    for k, v := range headers {
        w.Header().Set(k, v)
    }
    w.WriteHeader(respCode)
    if len(data) > 0 {
        w.Write(data)
    }
}

func (s *RESTRouter) jsonSerialize(dstruct interface{}, code int) ([]byte, int) {
    bytes, err := json.Marshal(&dstruct)
    if err != nil {
        s.logger.Error.Println(err)
        return []byte(fmt.Sprintf(`{"error": %s}`,err)), 500
    }
    return bytes, code
}

func (s *RESTRouter) writeJsonResponse(writer http.ResponseWriter, headers map[string]string, data interface{}, respCode int) int {
    bytes, code := s.jsonSerialize(data, respCode)

    hdrs := make(map[string]string)
    hdrs["Content-Type"] = "application/json"
    for k,v := range headers {
        hdrs[k] = v
    }
    s.writeHttpResponse(writer, hdrs, bytes, code)
    return code
}

func (s *RESTRouter) pathParts(path string) []string {
    parts := make([]string, 0)
    for _, v := range strings.Split(path, "/") {
        if v != "" {
            parts = append(parts, v)
        }
    }
    return parts
}

func (s *RESTRouter) Register(path string, hdlr RESTEndpointHandler) {
    parts := s.pathParts(path)
    s.logger.Debug.Printf("Registering path: %s%s\n", s.Prefix, path)

    // 0 reservded for root path.
    if len(parts) == len(s.handlerMap) {
        s.handlerMap = append(s.handlerMap, hdlr)
    } else if len(parts) > len(s.handlerMap) {
        tmap := make([]RESTEndpointHandler, len(parts)+1)
        for i, v := range s.handlerMap {
            tmap[i] = v
        }
        s.handlerMap = tmap
        s.handlerMap[len(parts)] = hdlr
    }
}

func (s *RESTRouter) runMethodHandler(r *http.Request, handlerIndex int, body map[string]interface{}, args... string) (interface{}, int) {
    var ( data interface{}; code int; )
    switch(r.Method) {
        case "GET":
            data, code = s.handlerMap[handlerIndex].GET(r, args...)
            break
        case "POST":
            data, code = s.handlerMap[handlerIndex].POST(r, args...)
            break
        case "PUT":
            data, code = s.handlerMap[handlerIndex].PUT(r, args...)
            break
        case "DELETE":
            data, code = s.handlerMap[handlerIndex].DELETE(r, args...)
            break
        case "PATCH":
            data, code = s.handlerMap[handlerIndex].PATCH(r, args...)
            break
        default:
            data = map[string]string{"error": "Invalid method"}
            code = 400
            break
    }
    return data, code
}

func (s *RESTRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // strip path before processing handler
    r.URL.Path = strings.TrimPrefix(r.URL.Path, s.Prefix)
    var ( data interface{}; code int; )
    parts := s.pathParts(r.URL.Path)
    if len(parts) < 0 || len(parts) > len(s.handlerMap) {
        data = map[string]string{"error": "Not found!"}
        code = 404
    } else {
        if r.Method == "OPTIONS" {
            w.Header().Set("Access-Control-Allow-Origin", s.defaultACLs)
            w.WriteHeader(200)
            return
        }
        if s.handlerMap[len(parts)] != nil {
            data, code = s.runMethodHandler(r, len(parts), nil, parts...)
        } else {
            data = map[string]string{"error": "Not found!"}
            code = 404
        }
    }
    s.writeJsonResponse(w, map[string]string{"Access-Control-Allow-Origin":s.defaultACLs},
                                                                            data, code)
    s.logger.Info.Printf("%s %d %s %s\n", r.Method, code, r.RequestURI, r.RemoteAddr)
}
