package magnet

import "reflect"

type TypeHook = func(Hook, reflect.Type) bool

type typeHooks struct {
	hooks []TypeHook
}

type Hook struct {
	m *Magnet
}

func (h Hook) Register(factory interface{}) {
	h.m.Register(factory).RecreateAlways()
}

func (h Hook) ValidateDeps(deps []reflect.Type) {
	h.m.validate(deps)
}

func (th *typeHooks) runHooks(m *Magnet, t reflect.Type) {
	for _, hook := range th.hooks {
		if hook(Hook{m}, t) {
			break
		}
	}
}

func (th *typeHooks) register(t TypeHook) {
	th.hooks = append(th.hooks, t)
}
