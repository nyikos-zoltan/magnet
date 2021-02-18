package magnet_test

import (
	"reflect"
	"testing"

	"github.com/nyikos-zoltan/magnet"
	"github.com/stretchr/testify/require"
)

type prepareA struct{ int }
type prepareB struct{ int }

func Test_Prepare(t *testing.T) {
	m := magnet.New()
	b := prepareB{0}
	m.Register(func(prepareA) *prepareB { return &b })
	c := m.NewCaller(func(b *prepareB) *prepareB { return b }, reflect.TypeOf(prepareA{}))
	require.NoError(t, m.Prepare())
	rv, err := c.Call(prepareA{0})
	require.NoError(t, err)
	require.Equal(t, &b, rv.Interface(0).(*prepareB))
}
