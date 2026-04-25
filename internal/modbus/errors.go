package modbus

import "fmt"

type ExceptionError struct {
	Code byte
}

func (e *ExceptionError) Error() string {
	return fmt.Sprintf("Modbus exception %s", exceptionName(e.Code))
}
