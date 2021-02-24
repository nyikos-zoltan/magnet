package magnet_test

import (
	"testing"

	"github.com/nyikos-zoltan/magnet"
	"github.com/stretchr/testify/require"
)

type args struct{ a []int }

func testF(inargs ...int) args {
	return args{a: inargs}
}

func Test_WithArgs_Empty(t *testing.T) {
	m := magnet.New()
	m.Register(magnet.WithArgs(testF))

	var a args
	require.NoError(t, m.Retrieve(&a))

	require.Len(t, a.a, 0)
}

func Test_WithArgs(t *testing.T) {
	m := magnet.New()
	m.Register(magnet.WithArgs(testF, 1, 2, 3))

	var a args
	require.NoError(t, m.Retrieve(&a))

	require.Len(t, a.a, 3)
	require.Equal(t, []int{1, 2, 3}, a.a)
}
