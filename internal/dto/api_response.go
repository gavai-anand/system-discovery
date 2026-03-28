package dto

type RequestDetails struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type RequestError struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []RequestDetails `json:"details"`
}

type SuccessResponse struct {
	TraceId      string      `json:"trace_id"`
	Success      bool        `json:"success"`
	ResponseCode int         `json:"response_code"`
	Message      string      `json:"message"`
	Data         interface{} `json:"data"`
}

type NotFoundResponse struct {
	TraceId      string                 `json:"trace_id"`
	Success      bool                   `json:"success"`
	ResponseCode string                 `json:"response_code"`
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data"`
}

type BadRequestResponse struct {
	TraceId      string           `json:"trace_id"`
	Success      bool             `json:"success"`
	ResponseCode string           `json:"response_code"`
	Message      string           `json:"message"`
	Details      []RequestDetails `json:"details,omitempty"` // List of field-specific validation errors
}

type ServerErrorResponse struct {
	TraceId      string                 `json:"trace_id"`
	Success      bool                   `json:"success"`
	ResponseCode string                 `json:"response_code"`
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data"`
}
