package magnet

import (
	"fmt"
	"reflect"
)

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

// Needs returns if the given instance of Magnet is required to build this node
func (n *Node) Needs(m *Magnet) bool {
	for _, v := range n.requires {
		if _, has := m.providerMap[v]; has {
			return true
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
