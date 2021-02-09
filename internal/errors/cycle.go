package errors

import (
	"fmt"
	"strings"
)

type CycleError struct {
	deps []string
}

func NewCycleError(deps []string) CycleError {
	return CycleError{deps}
}

func (e CycleError) Error() string {
	loop := append(e.deps, e.deps[0])
	return fmt.Sprintf("cycle found: %s", strings.Join(loop, "<-"))
}

func (e CycleError) Loop() (rv []string) {
	rv = make([]string, len(e.deps))
	copy(rv, e.deps)
	return
}
