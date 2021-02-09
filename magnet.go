package magnet

import (
	"fmt"
	"reflect"

	"github.com/nyikos-zoltan/magnet/internal/errors"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()
var magnetType = reflect.TypeOf((*Magnet)(nil))

func (m *Magnet) canBuildType(t reflect.Type) bool {
	return m.findNode(t) != nil
}

func (m *Magnet) validate(requires []reflect.Type) {
	for _, t := range requires {
		if !m.canBuildType(t) {
			panic(fmt.Sprintf("type %s cannot be constructed!", t))
		}
	}
}

func calculateRequiredFn(ftype reflect.Type) []reflect.Type {
	var rv []reflect.Type
	for i := 0; i < ftype.NumIn(); i++ {
		rv = append(rv, ftype.In(i))
	}
	return rv
}

func (m *Magnet) findNode(t reflect.Type) *Node {
	if node, has := m.providerMap[t]; has {
		return node
	}
	if m.parent != nil {
		return m.parent.findNode(t)
	}
	return nil
}

func (m *Magnet) call(fn interface{}, reqs []reflect.Type) ([]reflect.Value, error) {
	vals, err := m.BuildMany(reqs)
	if err != nil {
		return nil, err
	}
	return reflect.ValueOf(fn).Call(vals), nil
}

func (m *Magnet) makeLoopError(loop map[reflect.Type]bool) error {
	var deps []string
	var first reflect.Type
	for k := range loop {
		first = k
		break
	}

	n := first
	for n != nil {
		for _, r := range m.findNode(n).requires {
			deps = append(deps, n.Name())
			if r == first {
				n = nil
				break
			}

			if _, has := loop[r]; has {
				delete(loop, r)
				n = r
				break
			}
		}
	}
	return errors.NewCycleError(deps)
}

func (m *Magnet) dfs_visit(t reflect.Type, disc, finish map[reflect.Type]bool) error {
	node := m.findNode(t)
	if node == nil {
		return nil
	}
	disc[t] = true

	for _, v := range node.requires {
		if _, has := disc[v]; has {
			return m.makeLoopError(disc)
		}

		if _, has := finish[v]; !has {
			if err := m.dfs_visit(v, disc, finish); err != nil {
				return err
			}
		}
	}

	delete(disc, t)
	finish[t] = true

	return nil
}

func (m *Magnet) collectAllTypes(into *[]reflect.Type) {
	for k := range m.providerMap {
		*into = append(*into, k)
	}
	if m.parent != nil {
		m.parent.collectAllTypes(into)
	}
}

func (m *Magnet) dfs() error {
	disc := make(map[reflect.Type]bool)
	finish := make(map[reflect.Type]bool)

	var types []reflect.Type
	m.collectAllTypes(&types)

	for _, k := range types {
		if _, discovered := disc[k]; discovered {
			continue
		}
		if _, finished := finish[k]; finished {
			continue
		}

		if err := m.dfs_visit(k, disc, finish); err != nil {
			return err
		}
	}
	return nil
}

func (m *Magnet) detectCycles() {
	if !m.valid {
		if err := m.dfs(); err != nil {
			panic(err)
		}

	}
	m.valid = true
	if m.parent != nil {
		m.parent.detectCycles()
	}
}

func (m *Magnet) copyOwned(candidates []reflect.Type) {
	for _, v := range candidates {
		cNode := m.findNode(v)
		if cNode == nil {
			panic(fmt.Sprintf("type %s cannot be built!", v))
		}
		for _, node := range cNode.CollectDependencies(m) {
			if node.Needs(m) && node.owner != m {
				m.providerMap[node.provides] = node.cloneTo(m)
			}
		}
	}
}

// Reset clears all cached values in this instance
func (m *Magnet) Reset() {
	for k, v := range m.providerMap {
		if k != magnetType {
			v.value = reflect.Value{}
		}
	}

}

// Build uses the instance (and its parents) to resolve a specific value.
// This method will at this point create all dependencies of `t` to create a value of it.
func (m *Magnet) Build(t reflect.Type) (reflect.Value, error) {
	return m.findNode(t).Build(m)
}
func (m *Magnet) BuildMany(types []reflect.Type) ([]reflect.Value, error) {
	var vals []reflect.Value
	for _, req := range types {
		val, err := m.Build(req)
		if err != nil {
			return nil, err
		}
		vals = append(vals, val)
	}
	return vals, nil
}

// Magnet is a (hierarchic) IoC container that allows you to use argument injection.
type Magnet struct {
	parent      *Magnet
	providerMap map[reflect.Type]*Node
	valid       bool
	hooks       *typeHooks
}

func newMagnet(parent *Magnet) *Magnet {
	m := &Magnet{
		parent:      parent,
		providerMap: make(map[reflect.Type]*Node),
	}
	if parent != nil {
		m.hooks = parent.hooks
	} else {
		m.hooks = &typeHooks{}
	}
	m.providerMap[magnetType] = &Node{
		owner:    m,
		provides: magnetType,
		value:    reflect.ValueOf(m),
	}
	return m
}

// New creates a new instance of Magent.
func New() *Magnet {
	m := newMagnet(nil)
	m.RegisterTypeHook(derivedTypeHook)
	return m
}

// New creates a new child instance from m.
// Child instances will use factory methods defined in the parent instances as well.
func (m *Magnet) NewChild() *Magnet {
	return newMagnet(m)
}

func (m *Magnet) RegisterTypeHook(hook TypeHook) {
	m.hooks.register(hook)
}

func (m *Magnet) runHooks(types ...reflect.Type) {
	for _, t := range types {
		m.hooks.runHooks(m, t)
	}
}

type Factory struct {
	*Node
}

func (f *Factory) RecreateAlways() *Factory {
	f.forceRecreate = true
	return f
}

func (m *Magnet) RegisterMany(factories ...interface{}) {
	for _, factory := range factories {
		m.Register(factory)
	}
}

// Register adds a new factory to this instance.
// Factories are methods that have any number of arguments and return a single value and possible an error.
func (m *Magnet) Register(factory interface{}) *Factory {
	m.valid = false
	fval := reflect.ValueOf(factory)
	ftype := fval.Type()
	if ftype.NumOut() == 1 {
		reqs := calculateRequiredFn(ftype)
		m.runHooks(reqs...)
		n := &Node{
			owner:    m,
			provides: ftype.Out(0),
			requires: reqs,
			factory:  fval,
			fallible: false,
		}
		m.providerMap[ftype.Out(0)] = n
		return &Factory{n}
	}
	if ftype.NumOut() == 2 {
		if ftype.Out(1) == errorType {
			reqs := calculateRequiredFn(ftype)
			m.runHooks(reqs...)
			n := &Node{
				owner:    m,
				provides: ftype.Out(0),
				requires: reqs,
				factory:  fval,
				fallible: true,
			}
			m.providerMap[ftype.Out(0)] = n
			return &Factory{n}
		}
	}
	panic(fmt.Sprintf("invalid factory %s %s", ftype.Out(1), errorType))
}
