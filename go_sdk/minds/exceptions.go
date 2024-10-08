package minds

import "fmt"

// ObjectNotFound is raised when a requested object is not found.
type ObjectNotFound struct {
	Message string
}

func (e *ObjectNotFound) Error() string {
	return fmt.Sprintf("Object not found: %s", e.Message)
}

// ObjectNotSupported is raised when an action is not supported for the requested object.
type ObjectNotSupported struct {
	Message string
}

func (e *ObjectNotSupported) Error() string {
	return fmt.Sprintf("Object not supported: %s", e.Message)
}

// Forbidden is raised when an action is forbidden.
type Forbidden struct {
	Message string
}

func (e *Forbidden) Error() string {
	return fmt.Sprintf("Forbidden: %s", e.Message)
}

// Unauthorized is raised when authentication is required and has failed or has not been provided.
type Unauthorized struct {
	Message string
}

func (e *Unauthorized) Error() string {
	return fmt.Sprintf("Unauthorized: %s", e.Message)
}

// UnknownError is raised when an unknown error occurs.
type UnknownError struct {
	Message string
}

func (e *UnknownError) Error() string {
	return fmt.Sprintf("Unknown error: %s", e.Message)
}
