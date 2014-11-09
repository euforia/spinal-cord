package web

import(
    "encoding/json"
    "io/ioutil"
    "net/http"
)

type RESTEndpointHandler interface{
    GET(r *http.Request, args... string) (interface{}, int)
    PATCH(r *http.Request, args... string) (interface{}, int)
    DELETE(r *http.Request, args... string) (interface{}, int)
    POST(r *http.Request, args... string) (interface{}, int)
    PUT(r *http.Request, args... string) (interface{}, int)
}

type DefaultEndpointHandler struct{}
func (d *DefaultEndpointHandler) GET(r *http.Request, args... string) (interface{}, int) {
    return map[string]string{"error": "Invalid method"}, 400
}
func (d *DefaultEndpointHandler) PATCH(r *http.Request, args... string) (interface{}, int) {
    return map[string]string{"error": "Invalid method"}, 400
}
func (d *DefaultEndpointHandler) DELETE(r *http.Request, args... string) (interface{}, int) {
    return map[string]string{"error": "Invalid method"}, 400
}
func (d *DefaultEndpointHandler) POST(r *http.Request, args... string) (interface{}, int) {
    return map[string]string{"error": "Invalid method"}, 400
}
func (d *DefaultEndpointHandler) PUT(r *http.Request, args... string) (interface{}, int) {
    return map[string]string{"error": "Invalid method"}, 400
}

func (d *DefaultEndpointHandler) JsonBody(r *http.Request) (map[string]interface{}, error) {
    body, err := ioutil.ReadAll(r.Body);
    if err != nil {
        return nil, err
    }
    var tdata map[string]interface{}
    err = json.Unmarshal(body , &tdata)
    if err != nil {
        return nil, err
    }
    return tdata, nil
}
