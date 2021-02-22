package magnet

import (
	"fmt"
	"reflect"
	"sort"

	stdErrors "errors"

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
			panic(errors.NewCannotBeBuiltErr(t))
		}
	}
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

func (m *Magnet) makeLoopError(loop map[reflect.Type]*Node) error {
	var deps []string
	for _, n := range loop {
		deps = append(deps, n.provides.Name())
	}
	return errors.NewCycleError(deps)
}

func (m *Magnet) dfs_visit(t reflect.Type, disc map[reflect.Type]*Node, finish map[reflect.Type]bool) error {
	node := m.findNode(t)
	if node == nil {
		return nil
	}
	disc[t] = node

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

type byName []reflect.Type

func (t byName) Len() int           { return len(t) }
func (t byName) Less(i, j int) bool { return t[i].Name() < t[j].Name() }
func (t byName) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

func (m *Magnet) collectAllTypes(into *[]reflect.Type) {
	for k := range m.providerMap {
		*into = append(*into, k)
	}
	if m.parent != nil {
		m.parent.collectAllTypes(into)
	}
	sort.Sort(byName(*into))
}

func (m *Magnet) dfs() error {
	disc := make(map[reflect.Type]*Node)
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

func (m *Magnet) NeedsAny(dep []reflect.Type) []*Node {
	var nodes []*Node
	for _, node := range m.providerMap {
		if node.NeedsAny(dep) {
			nodes = append(nodes, node)
		}
	}
	if m.parent != nil {
		nodes = append(nodes, m.parent.NeedsAny(dep)...)
	}
	return nodes
}

// Reset clears all cached values in this instance
func (m *Magnet) Reset() {
	for k, v := range m.providerMap {
		if k != magnetType {
			v.Reset()
		}
	}

}

var UnknownTypeErr = stdErrors.New("unknown type")

// Build uses the instance (and its parents) to resolve a specific value.
// This method will at this point create all dependencies of `t` to create a value of it.
func (m *Magnet) Build(t reflect.Type) (reflect.Value, error) {
	if node := m.findNode(t); node != nil {
		return m.findNode(t).Build(m)
	} else {
		return reflect.Value{}, fmt.Errorf("%s is %w", t, UnknownTypeErr)
	}
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

type Plugin = func(*Magnet)

// New creates a new instance of Magent.
func New(plugins ...Plugin) *Magnet {
	m := newMagnet(nil)
	m.RegisterTypeHook(derivedTypeHook)
	for _, plugin := range plugins {
		plugin(m)
	}
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
	node, err := NewNode(factory, m)
	if err != nil {
		panic(err)
	}
	m.runHooks(node.requires...)
	m.providerMap[node.provides] = node
	return &Factory{node}
}
