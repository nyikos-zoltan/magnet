package magnet_test

import (
	"testing"

	"github.com/nyikos-zoltan/magnet"
	"github.com/stretchr/testify/require"
)

func Test_Manget_Call(t *testing.T) {
	m := magnet.New()
	type a struct{ int }
	m.Value(a{1})
	rv, err := m.Call(func(a a) int { return a.int })
	require.NoError(t, err)
	require.Equal(t, 1, rv.Interface(0))
}
