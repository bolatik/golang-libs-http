package http

type Response interface {
	Headers() map[string]string
	SetHeader(key, value string)
	GetHeader(key string) string

	Response() interface{}
	StatusCode() int
}

type response struct {
	Status     int
	Data       interface{}
	HeaderData map[string]string
}

func (e *response) Response() interface{} {
	return e.Data
}

func (e *response) StatusCode() int {
	return e.Status
}

func (e *response) Headers() map[string]string {
	return e.HeaderData
}

func (r *response) SetHeader(key, value string) {
	r.HeaderData[key] = value
}

func (r *response) GetHeader(key string) string {
	return r.HeaderData[key]
}

func NewResponse(status int, data interface{}, headers map[string]string) Response {
	return &response{status, data, headers}
}
