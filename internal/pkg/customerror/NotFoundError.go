package customerror

// NotFoundError is rates not found error
type NotFoundError struct {
	message string
}

// Error returns error message
func (m *NotFoundError) Error() string {
	return m.message
}

// NewNotFoundError error constructor
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		message: message,
	}
}
