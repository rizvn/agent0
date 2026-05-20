package util

import (
	"fmt"
	"runtime"
)

// DetailedError return error with a message
// and wraps the original error
// with additional context about where the error occurred.
func DetailedError(message string, wrapErr error) error {
	// get refenece to caller function
	pc, _, line, _ := runtime.Caller(1)

	// get function name of the caller
	funcName := runtime.FuncForPC(pc).Name()

	return fmt.Errorf("%s",
		fmt.Sprintf("\nError at:%s Line:%d\nMessage:%s\n%v",
			funcName, line, message, wrapErr))
}
