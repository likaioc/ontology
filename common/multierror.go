package common

import (
	"bytes"
	"fmt"
)

type Errors []error

func (e Errors) Err() error {

	bAllNilFlag := true
	for er := range e {
		if er != 0 {
			bAllNilFlag = false
		}
	}

	if bAllNilFlag {
		return nil
	}

	return &MultiError{Errors: e}
}

type MultiError struct {
	Errors Errors
}

func (m *MultiError) Error() string {
	var buf bytes.Buffer

	if len(m.Errors) == 1 {
		buf.WriteString("1 error: ")
	} else{
		fmt.Fprintf(&buf, "%d errors: ", len(m.Errors))
	}

	for i, err := range m.Errors {
		if i != 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(err.Error())
	}

	return buf.String()
}


