package magnet_test

import (
	"testing"

	"github.com/nyikos-zoltan/magnet"
	"github.com/stretchr/testify/require"
)

type valueA struct{ int }

func Test_Value(t *testing.T) {
	req := require.New(t)
	m := magnet.New()
	m.Value(valueA{1})
	c := m.NewCaller(func(a valueA) int { return a.int })
	rv, err := c.Call()
	req.NoError(err)
	req.EqualValues(1, rv.Interface(0))
}
