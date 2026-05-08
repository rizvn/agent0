package util

import (
	"fmt"
	"runtime"
)

type Err struct {
	message string
}

func (e *Err) Error() string {
	return e.message
}

// NewErr Returns a new error with message contaning the funcName and line number
// of where the error occured, this is determined from caller of this method
func NewErr(message string, wrapErr error) *Err {
	// get refenece to caller function
	pc, _, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()

	e := &Err{}
	e.message = fmt.Sprintf("\nError at:%s Line:%d\nMessage:%s\n%v", funcName, line, message, wrapErr)
	return e
}
