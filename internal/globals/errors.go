package globals

import "fmt"

// ExitError carries an exit code alongside an error message.
type ExitError struct {
	Code int
	Msg  string
}

func (e *ExitError) Error() string { return e.Msg }

// Exitf returns an ExitError with the given code and formatted message.
func Exitf(code int, format string, args ...any) error {
	return &ExitError{Code: code, Msg: fmt.Sprintf(format, args...)}
}
