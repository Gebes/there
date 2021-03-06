package there

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type HttpRequest struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter

	Method      string
	Body        *BodyReader
	Params      *BasicReader
	Headers     *BasicReader
	RouteParams *RouteParamReader
}

func NewHttpRequest(responseWriter http.ResponseWriter, request *http.Request) HttpRequest {
	paramReader := BasicReader(request.URL.Query())
	headerReader := BasicReader(request.Header)
	routeParamReader := RouteParamReader(MapString{})
	return HttpRequest{
		Request:        request,
		ResponseWriter: responseWriter,
		Method:         request.Method,
		Body:           &BodyReader{request: request},
		Params:         &paramReader,
		Headers:        &headerReader,
		RouteParams:    &routeParamReader,
	}
}

func (r *HttpRequest) Context() context.Context {
	return r.Request.Context()
}

func (r *HttpRequest) WithContext(ctx context.Context) {
	*r.Request = *r.Request.WithContext(ctx)
}

//BodyReader reads the body and unmarshal it to the specified destination
type BodyReader struct {
	request *http.Request
}

func (read BodyReader) BindJson(dest any) error {
	return read.bind(dest, json.Unmarshal)
}

func (read BodyReader) BindXml(dest any) error {
	return read.bind(dest, xml.Unmarshal)
}

func (read BodyReader) bind(dest any, formatter func(data []byte, v any) error) error {
	body, err := read.ToBytes()
	if err != nil {
		return err
	}
	err = formatter(body, dest)
	return err
}

func (read BodyReader) ToString() (string, error) {
	data, err := read.ToBytes()
	if err != nil {
		return "", err
	}
	return string(data), err
}

func (read BodyReader) ToBytes() ([]byte, error) {
	data, err := ioutil.ReadAll(read.request.Body)
	defer read.request.Body.Close()
	if err != nil {
		return nil, err
	}
	return data, nil
}

//BasicReader reads http params
type BasicReader map[string][]string

func (reader BasicReader) Has(key string) bool {
	_, ok := reader.GetSlice(key)
	return ok
}

func (reader BasicReader) GetDefault(key, defaultValue string) string {
	s, ok := reader.Get(key)
	if !ok {
		return defaultValue
	}
	return s
}

func (reader BasicReader) Get(key string) (string, bool) {
	list, ok := reader.GetSlice(key)
	if !ok {
		return "", false
	}
	return list[0], true
}

func (reader BasicReader) GetSlice(key string) ([]string, bool) {
	list, ok := reader[key]
	if !ok || len(list) == 0 {
		return nil, false
	}
	return list, true
}

//RouteParamReader reads dynamic route params
type RouteParamReader map[string]string

func (reader RouteParamReader) Has(key string) bool {
	_, ok := reader[key]
	return ok
}

func (reader RouteParamReader) GetDefault(key, defaultValue string) string {
	s, ok := reader.Get(key)
	if !ok {
		return defaultValue
	}
	return s
}

func (reader RouteParamReader) Get(key string) (string, bool) {
	v, ok := reader[key]
	return v, ok
}
