package magnet

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_findPath(t *testing.T) {
	type testA struct{}
	type testB struct{}
	type testC struct{}
	typeA := reflect.TypeOf(testA{})
	typeB := reflect.TypeOf(testB{})
	typeC := reflect.TypeOf(testC{})
	m := New()
	m.RegisterMany(
		func() testA { return testA{} },
		func(testA) testB { return testB{} },
		func(testB) testC { return testC{} },
	)
	pth := m.findPath(
		typeC,
		typeB,
	)
	require.Equal(t,
		map[reflect.Type]bool{typeC: true, typeB: true},
		pth,
	)

	pth = m.findPath(
		typeC,
		typeA,
	)
	require.Equal(t,
		map[reflect.Type]bool{typeC: true, typeB: true, typeA: true},
		pth,
	)
}

func Test_findPathAll(t *testing.T) {
	type test1 struct{}
	type test2A struct{}
	type test2B struct{}
	type test3A struct{}
	type test3B struct{}
	type test3C struct{}
	type test3D struct{}
	type leaf struct{}

	test1Type := reflect.TypeOf(test1{})
	test2AType := reflect.TypeOf(test2A{})
	test2BType := reflect.TypeOf(test2B{})
	test3AType := reflect.TypeOf(test3A{})
	test3BType := reflect.TypeOf(test3B{})
	test3CType := reflect.TypeOf(test3C{})
	leafType := reflect.TypeOf(leaf{})

	m := New()
	m.RegisterMany(
		func(test2A, test2B) test1 { return test1{} },

		func(test3A, test3B) test2A { return test2A{} },
		func(test3C, test3D) test2B { return test2B{} },

		func(leaf) test3A { return test3A{} },
		func(leaf) test3B { return test3B{} },
		func(leaf) test3C { return test3C{} },
		func() test3D { return test3D{} },

		func() leaf { return leaf{} },
	)

	require.Equal(t,
		map[reflect.Type]bool{test1Type: true, test2AType: true, test2BType: true, test3AType: true, test3BType: true, test3CType: true, leafType: true},
		m.findPath(test1Type, leafType),
	)
}
