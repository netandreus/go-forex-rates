package custom_errors

// BadRequestError represents bad RatesRequest error
type BadRequestError struct {
	message string
}

// Error returns error message
func (m *BadRequestError) Error() string {
	return m.message
}

// NewBadRequestError error constructor
func NewBadRequestError(message string) *BadRequestError {
	return &BadRequestError{
		message: message,
	}
}
