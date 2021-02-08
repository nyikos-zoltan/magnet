package magnet

import (
	"reflect"

	"github.com/labstack/echo/v4"
)

func validateEchoHandlerFn(fntype reflect.Type) {
	if fntype.NumOut() != 1 {
		panic("EchoHandler methods need to return an `error`")
	}
	if fntype.Out(0) != errorType {
		panic("EchoHandler methods need to return an `error`")
	}
}

var ctxType = reflect.TypeOf((*echo.Context)(nil)).Elem()
var handlerType = reflect.TypeOf((*echo.HandlerFunc)(nil)).Elem()

// EchoHandler creates a new echo.HandlerFunc that injects the required values
func (m *Magnet) EchoHandler(fn interface{}) func(echo.Context) error {
	m.detectCycles()
	fntype := reflect.TypeOf(fn)
	validateEchoHandlerFn(fntype)
	caller := m.NewCaller(fn, ctxType)

	return func(ctx echo.Context) error {
		rv, err := caller.Call(ctx)
		if err != nil {
			return err
		}
		return rv.Error(0)
	}
}

func (m *Magnet) EchoMiddleware(fn interface{}) echo.MiddlewareFunc {
	m.detectCycles()
	caller := m.NewCaller(fn, handlerType)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		rv, _ := caller.Call(next)
		return rv.Interface(0).(echo.HandlerFunc)
	}
}
