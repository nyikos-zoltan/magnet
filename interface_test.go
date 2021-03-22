package magnet_test

import (
	"testing"

	"github.com/nyikos-zoltan/magnet"
	"github.com/stretchr/testify/require"
)

type interfaceA interface {
	Method() int
}

func (a *valueA) Method() int { return a.int }

func Test_Interface(t *testing.T) {
	req := require.New(t)
	m := magnet.New(magnet.InterfacePlugin)
	m.Value(&valueA{1})
	c := m.NewCaller(func(a interfaceA) int { return a.Method() })
	rv, err := c.Call()
	req.NoError(err)
	req.EqualValues(1, rv.Interface(0))
}
