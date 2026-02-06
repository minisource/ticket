package handlers

// Response represents a standardized API response for swagger docs
// This mirrors the go-common response.Response struct
type Response struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      *ErrorInfo  `json:"error,omitempty"`
	Meta       *Meta       `json:"meta,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	TraceID    string      `json:"traceId,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	Field      string            `json:"field,omitempty"`
	Validation []ValidationError `json:"validation,omitempty"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Meta contains metadata about the response
type Meta struct {
	RequestID string `json:"requestId,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// Pagination contains pagination information
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}
