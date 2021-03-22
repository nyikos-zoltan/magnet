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
var errType = reflect.TypeOf((*error)(nil)).Elem()

func validateEchoErrorHandlerFn(fntype reflect.Type) {
	hasError := false
	hasCtx := false

	for i := 0; i < fntype.NumIn(); i++ {
		if fntype.In(i) == ctxType {
			hasCtx = true
		} else if fntype.In(i) == errType {
			hasError = true
		}
	}

	if !hasError || !hasCtx {
		panic("EchoErrorHandler methods need to take at least an error and the echo.Context")
	}
}

// EchoHandler creates a new echo.HandlerFunc that injects the required values
func (m *Magnet) EchoHandler(fn interface{}) func(echo.Context) error {
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
	caller := m.NewCaller(fn, handlerType)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		rv, _ := caller.Call(next)
		return rv.Interface(0).(echo.HandlerFunc)
	}
}

func (m *Magnet) EchoErrorHandler(fn interface{}) echo.HTTPErrorHandler {
	validateEchoErrorHandlerFn(reflect.TypeOf(fn))
	caller := m.NewCaller(fn, ctxType, errType)
	return func(e error, c echo.Context) {
		_, _ = caller.Call(c, e)
	}
}
