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
		return rv.Error(0)
	}
}
