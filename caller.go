package magnet

import "reflect"

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
	return cr.vals[0].String()
}

// Caller prepares a method and a number of extra types to be repeatedly called later.
type Caller struct {
	fn         interface{}
	fntype     reflect.Type
	reqs       []reflect.Type
	extraNodes []*Node
	owner      *Magnet
}

// NewCaller creates a new caller from a method (fn) and some predefined types (extraTypes).
// It's important to note that the order of the `extraTypes` here and the specific arguments in the `Caller.Call` method matter.
func (m *Magnet) NewCaller(fn interface{}, extraTypes ...reflect.Type) *Caller {
	child := m.NewChild()
	fntype := reflect.TypeOf(fn)
	reqs := calculateRequiredFn(fntype)
	m.runHooks(reqs...)
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
	child.copyOwned(reqs)

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
