package magnet

import (
	"fmt"
	"reflect"
)

func calculateRequiredFn(ftype reflect.Type) []reflect.Type {
	var rv []reflect.Type
	for i := 0; i < ftype.NumIn(); i++ {
		rv = append(rv, ftype.In(i))
	}
	return rv
}

func newRawValueNode(val reflect.Value, owner *Magnet) *Node {
	vtype := val.Type()
	return &Node{
		owner:    owner,
		provides: vtype,
		value:    val,
	}
}

func NewValueNode(v interface{}, owner *Magnet) *Node {
	val := reflect.ValueOf(v)
	return newRawValueNode(val, owner)
}

func NewEmptyNode(t reflect.Type, owner *Magnet) *Node {
	return &Node{
		owner:    owner,
		provides: t,
	}
}

func NewNode(factory interface{}, owner *Magnet) (*Node, error) {
	fval := reflect.ValueOf(factory)
	ftype := fval.Type()
	reqs := calculateRequiredFn(ftype)
	if ftype.NumOut() == 1 {
		return &Node{
			owner:    owner,
			provides: ftype.Out(0),
			requires: reqs,
			factory:  fval,
			fallible: false,
		}, nil
	}
	if ftype.NumOut() == 2 {
		if ftype.Out(1) == errorType {
			return &Node{
				owner:    owner,
				provides: ftype.Out(0),
				requires: reqs,
				factory:  fval,
				fallible: true,
			}, nil
		}
	}
	return nil, fmt.Errorf("invalid factory %s", ftype)
}

// Node describes a factory method, its inputs and outputs
type Node struct {
	owner         *Magnet
	provides      reflect.Type
	requires      []reflect.Type
	factory       reflect.Value
	fallible      bool
	value         reflect.Value
	forceRecreate bool
}

func (n *Node) cloneTo(m *Magnet) *Node {
	return &Node{
		owner:         m,
		provides:      n.provides,
		requires:      n.requires,
		factory:       n.factory,
		fallible:      n.fallible,
		value:         n.value,
		forceRecreate: n.forceRecreate,
	}
}

var valueType = reflect.TypeOf((*reflect.Value)(nil)).Elem()

func unwrapValue(v reflect.Value) reflect.Value {
	if v.Type() == valueType {
		return v.Interface().(reflect.Value)
	} else {
		return v
	}
}

func (n *Node) Reset() {
	if n.factory != (reflect.Value{}) {
		n.value = reflect.Value{}
	}
}

// Build uses the factory method to generate its value, then stores it for later use
func (n *Node) Build(m *Magnet) (reflect.Value, error) {
	if n.value == (reflect.Value{}) || n.forceRecreate {
		params, err := m.BuildMany(n.requires)
		if err != nil {
			return reflect.Value{}, err
		}
		if n.factory == (reflect.Value{}) {
			return reflect.Value{}, fmt.Errorf("failed to produce %s, missing factory", n.provides)
		}
		rv := n.factory.Call(params)
		if n.fallible {
			err, _ := rv[1].Interface().(error)
			n.value = unwrapValue(rv[0])
			return rv[0], err
		} else {
			n.value = unwrapValue(rv[0])
			return n.value, nil
		}
	}
	return n.value, nil
}

func (n *Node) NeedsAny(types []reflect.Type) bool {
	for _, v := range n.requires {
		for _, t := range types {
			if v == t {
				return true
			}
		}
	}
	return false
}

// CollectDependencies collects all the nodes that are required by this one
func (n *Node) CollectDependencies(m *Magnet) []*Node {
	var ret []*Node
	ret = append(ret, n)
	keys := make(map[*Node]bool)
	for _, v := range n.requires {
		n = m.findNode(v)
		if n == nil {
			panic(fmt.Sprintf("type %s cannot be built!", v))
		}
		for _, v := range n.CollectDependencies(m) {
			if _, has := keys[v]; !has {
				keys[v] = true
				ret = append(ret, v)
			}
		}
	}
	return ret
}
