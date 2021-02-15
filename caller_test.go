package magnet_test

import (
	"reflect"
	"testing"

	"github.com/nyikos-zoltan/magnet"
	"github.com/stretchr/testify/require"
)

type a struct{ int }

func Test_CallerOverride(t *testing.T) {
	m := magnet.New()
	m.Register(func() a { return a{0} })

	_, err := m.NewCaller(func(v a) {
	}).Call()
	require.NoError(t, err)

	c := m.NewCaller(func(v a) int {
		return v.int
	}, reflect.TypeOf(a{}))

	rv, err := c.Call(a{1})
	require.NoError(t, err)
	require.Equal(t, 1, rv.Interface(0))
}
