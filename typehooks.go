package magnet

import "reflect"

type TypeHook = func(Hook, reflect.Type) bool

type typeHooks struct {
	hooks []TypeHook
}

type Hook struct {
	m *Magnet
}

type HookResult struct {
}

func (h Hook) RegisterNewType(requires []reflect.Type, provides reflect.Type, factory interface{}) {
	h.m.providerMap[provides] = &Node{
		requires:      requires,
		provides:      provides,
		owner:         h.m,
		factory:       reflect.ValueOf(factory),
		forceRecreate: true,
	}
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
