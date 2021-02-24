package magnet_test

import (
	"testing"

	"github.com/nyikos-zoltan/magnet"
	"github.com/stretchr/testify/require"
)

type autoRefA struct{ int }

func Test_AutoRef_Deref(t *testing.T) {
	m := magnet.New()
	m.Register(func() *autoRefA { return &autoRefA{1} }).AutoRef()

	var a autoRefA
	var aptr *autoRefA
	require.NoError(t, m.Retrieve(&a))
	require.NoError(t, m.Retrieve(&aptr))
	require.EqualValues(t, 1, a.int)
	require.EqualValues(t, 1, aptr.int)
}

func Test_AutoRef_Ref(t *testing.T) {
	m := magnet.New()
	m.Register(func() autoRefA { return autoRefA{1} }).AutoRef()

	var a autoRefA
	var aptr *autoRefA
	require.NoError(t, m.Retrieve(&a))
	require.NoError(t, m.Retrieve(&aptr))
	require.EqualValues(t, 1, a.int)
	require.EqualValues(t, 1, aptr.int)
}
