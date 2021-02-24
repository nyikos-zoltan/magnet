package magnet_test

import (
	"testing"

	"github.com/nyikos-zoltan/magnet"
	"github.com/stretchr/testify/require"
)

type retrieveA struct{ int }
type retrieveD struct {
	magnet.Derived
	A retrieveA
}

func Test_Retrieve_Derive(t *testing.T) {
	m := magnet.New()
	m.Register(func() retrieveA { return retrieveA{1} })
	var d retrieveD
	require.NoError(t, m.Retrieve(&d))
	require.EqualValues(t, 1, d.A.int)
}
