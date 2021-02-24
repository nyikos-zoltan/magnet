package magnet

import (
	"reflect"
)

type CallResults struct{ vals []reflect.Value }

func (cr CallResults) Len() int {
	return len(cr.vals)
}

func (cr CallResults) Interface(n int) interface{} {
	return cr.vals[n].Interface()
}

func (cr CallResults) Error(n int) error {
	rv := cr.vals[n]
	if rv.IsZero() {
		return nil
	} else {
		return rv.Interface().(error)
	}
}

func (cr CallResults) String(n int) string {
	return cr.vals[n].String()
}

func (cr CallResults) Value(n int) reflect.Value {
	return cr.vals[n]
}

// Caller prepares a method and a number of extra types to be repeatedly called later.
type Caller struct {
	fn         interface{}
	fntype     reflect.Type
	reqs       []reflect.Type
	extraNodes []*Node
	owner      *Magnet
}

func (m *Magnet) findPath(from reflect.Type, to reflect.Type) map[reflect.Type]bool {
	path := make(map[reflect.Type]bool)
	path[from] = true
	fromNode := m.findNode(from)
	if fromNode == nil {
		return nil
	}
	for _, req := range fromNode.requires {
		if req != to {
			rest := m.findPath(req, to)
			for k := range rest {
				path[k] = true
			}
		} else {
			path[from] = true
			path[to] = true
		}
	}
	if len(path) > 1 {
		return path
	} else {
		return nil
	}
}

func findOverrides(m *Magnet, reqs []reflect.Type, extraTypes map[reflect.Type]bool) map[*Node]bool {
	overrides := make(map[*Node]bool)
	for _, req := range reqs {
		for extra := range extraTypes {
			path := m.findPath(req, extra)
			for t := range path {
				node := m.findNode(t)
				overrides[node] = true
			}
		}
	}
	return overrides
}

// NewCaller creates a new caller from a method (fn) and some predefined types (extraTypes).
// It's important to note that the order of the `extraTypes` here and the specific arguments in the `Caller.Call` method matter.
func (m *Magnet) NewCaller(fn interface{}, extraTypes ...reflect.Type) *Caller {
	child := m.NewChild()
	fntype := reflect.TypeOf(fn)
	reqs := calculateRequiredFn(fntype)
	child.runHooks(reqs...)
	var extraNodes []*Node
	for _, extraType := range extraTypes {
		node := &Node{
			provides: extraType,
			owner:    child,
		}
		child.providerMap[extraType] = node
		extraNodes = append(extraNodes, node)
	}
	child.detectCycles()
	child.validate(reqs)

	extraTypeMap := make(map[reflect.Type]bool)
	for _, extra := range extraTypes {
		extraTypeMap[extra] = true
	}

	for n := range findOverrides(child, reqs, extraTypeMap) {
		if n.owner != child {
			node := n.cloneTo(child)
			child.providerMap[node.provides] = node
		}
	}

	return &Caller{
		fn:         fn,
		owner:      child,
		reqs:       reqs,
		fntype:     fntype,
		extraNodes: extraNodes,
	}
}

// Call uses the extra arguments to call the method contained in this caller.
// It's important to note that the order of the `extras` here and the argument types in the `NewCaller` method matter. This method will panic if the types are incompatible, unfortunately as of now there is no way to avoid this panic, you should always make sure that these calls are correct.
func (c *Caller) Call(extras ...interface{}) (*CallResults, error) {
	c.owner.Reset()
	for idx, extra := range extras {
		c.extraNodes[idx].value = reflect.ValueOf(extra)
	}
	rv, err := c.owner.call(c.fn, c.reqs)
	if err != nil {
		return nil, err
	}

	return &CallResults{rv}, nil
}
