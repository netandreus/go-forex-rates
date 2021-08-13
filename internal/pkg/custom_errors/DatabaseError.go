package custom_errors

// DatabaseError represents database error
type DatabaseError struct {
	message string
}

// Error returns error message
func (m *DatabaseError) Error() string {
	return m.message
}

// NewDatabaseError error constructor
func NewDatabaseError(message string) *DatabaseError {
	return &DatabaseError{
		message: message,
	}
}
