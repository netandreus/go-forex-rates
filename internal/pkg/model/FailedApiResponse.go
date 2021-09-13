package model

// ApiError represents error in API
type ApiError struct {
	// Error code
	Code int `json:"code"`

	// Error message
	Info string `json:"info"`
}

// FailedApiResponse response, when error occurred
type FailedApiResponse struct {
	// Success Returns true or false depending on whether or not your API request has succeeded.
	Success bool `json:"success"`

	// Code and error message
	Error ApiError `json:"error"`
}

// NewFailedApiResponse constructor
func NewFailedApiResponse(code int, info string) *FailedApiResponse {
	return &FailedApiResponse{
		Success: false,
		Error:   ApiError{Code: code, Info: info},
	}
}

// GetSuccess returns Success value
func (a *FailedApiResponse) GetSuccess() bool {
	return a.Success
}

// SetSuccess sets Success value
func (a *FailedApiResponse) SetSuccess(success bool) *FailedApiResponse {
	a.Success = success
	return a
}

// GetError returns Error value
func (a *FailedApiResponse) GetError() ApiError {
	return a.Error
}

// SetError sets Error value
func (a *FailedApiResponse) SetError(e ApiError) *FailedApiResponse {
	a.Error = e
	return a
}
