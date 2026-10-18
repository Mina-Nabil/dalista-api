package responses

// Pagination describes a paged result set's metadata.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginationResponse wraps a data payload with pagination metadata.
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"meta"`
}

// ErrorResponse is the generic API error envelope.
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// ValidationErrorResponse is the envelope for request-validation failures.
type ValidationErrorResponse struct {
	Errors []string `json:"errors" example:"phone is invalid,email is invalid"`
}
