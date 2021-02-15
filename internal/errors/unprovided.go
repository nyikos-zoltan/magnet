package errors

import (
	"fmt"
	"reflect"
)

type CannotBeBuiltErr struct {
	missing reflect.Type
}

func (e CannotBeBuiltErr) Error() string {
	return fmt.Sprintf("type %s cannot be constructed", e.missing)
}

func NewCannotBeBuiltErr(missing reflect.Type) CannotBeBuiltErr {
	return CannotBeBuiltErr{missing}
}
