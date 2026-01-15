package response

// Meta represents the meta information in API responses
type Meta struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Count   int    `json:"count,omitempty"`
}

// SingleDataResponse represents a response with single data item
type SingleDataResponse struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

// MultipleDataResponse represents a response with multiple data items
type MultipleDataResponse struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

// NewMeta creates a new meta object
func NewMeta(code int, status, message string) Meta {
	return Meta{
		Code:    code,
		Status:  status,
		Message: message,
	}
}

// NewMetaWithCount creates a new meta object with count
func NewMetaWithCount(code int, status, message string, count int) Meta {
	return Meta{
		Code:    code,
		Status:  status,
		Message: message,
		Count:   count,
	}
}

// NewSingleDataResponse creates a new single data response
func NewSingleDataResponse(meta Meta, data interface{}) SingleDataResponse {
	return SingleDataResponse{
		Meta: meta,
		Data: data,
	}
}

// NewMultipleDataResponse creates a new multiple data response
func NewMultipleDataResponse(meta Meta, data interface{}) MultipleDataResponse {
	return MultipleDataResponse{
		Meta: meta,
		Data: data,
	}
}

// ErrorResponse represents a response for errors
func ErrorResponse(code int, status, message string) SingleDataResponse {
	meta := NewMeta(code, status, message)
	return NewSingleDataResponse(meta, nil)
}
