package magnet

import (
	"reflect"

	workers "github.com/digitalocean/go-workers2"
)

// WorkerFunction creates workers.JobFunc that injects the required values
func (m *Magnet) WorkerFunction(fn interface{}) func(msg *workers.Msg) error {
	caller := m.NewCaller(fn, reflect.TypeOf(&workers.Msg{}))
	return func(msg *workers.Msg) error {
		rv, err := caller.Call(msg)
		if err != nil {
			return err
		}
		if err, ok := rv[0].Interface().(error); ok {
			return err
		}
		return nil
	}
}
