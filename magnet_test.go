package magnet_test

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nyikos-zoltan/magnet"
	magnetErrors "github.com/nyikos-zoltan/magnet/internal/errors"
	"github.com/stretchr/testify/require"
)

type A struct {
}

type I interface {
	M()
}

type implI struct {
	int
}

func (implI) M() {
}

type DerivedStruct struct {
	magnet.Derived
	InjI I
}

type AnonDerivedStruct struct {
	magnet.Derived
	I
}

func recoverPanic(f func()) interface{} {
	var msg interface{}
	func() {
		defer func() {
			msg = recover()
		}()
		f()
	}()
	return msg
}

func checkCycleError(t *testing.T, expectedCycle []string, panicMsg interface{}) {
	t.Helper()
	err, ok := panicMsg.(error)
	require.True(t, ok)
	var ce magnetErrors.CycleError
	require.True(t, errors.As(err, &ce))
	loop := ce.Loop()
	sort.Strings(loop)
	sort.Strings(expectedCycle)
	require.Equal(t, expectedCycle, loop)
}

func Test_Magnet(t *testing.T) {
	t.Run("ok - simple", func(t *testing.T) {
		m := magnet.New()
		m.Register(func() A {
			return A{}
		})

		var injA *A
		rv, err := m.NewCaller(func(a A) {
			injA = &a
		}).Call()
		require.NoError(t, err)
		require.Zero(t, rv.Len())
		require.NotNil(t, injA)
	})

	t.Run("ok - derived struct", func(t *testing.T) {
		m := magnet.New()
		m.RegisterMany(
			func() I {
				return &implI{1}
			},
		)
		var injD *DerivedStruct
		_, err := m.NewCaller(func(d DerivedStruct) {
			injD = &d
		}).Call()
		require.NoError(t, err)
		require.NotNil(t, injD)
		require.NotZero(t, injD.InjI)
	})

	t.Run("ok - derived struct with anonymous prop", func(t *testing.T) {
		m := magnet.New()
		m.RegisterMany(
			func() I {
				return &implI{1}
			},
		)
		var injD *AnonDerivedStruct
		_, err := m.NewCaller(func(d AnonDerivedStruct) {
			injD = &d
		}).Call()
		require.NoError(t, err)
		require.NotNil(t, injD)
		require.NotZero(t, injD.I)
	})

	t.Run("ok - interface factory", func(t *testing.T) {
		m := magnet.New()
		m.Register(func() I {
			return &implI{1}
		})
		var injI I
		_, err := m.NewCaller(func(i I) {
			injI = i
		}).Call()
		require.NoError(t, err)
		require.NotNil(t, injI)
	})

	t.Run("ok - reset", func(t *testing.T) {
		m := magnet.New()
		count := 0
		ctx := echo.New().NewContext(nil, nil)
		m.Register(func() I {
			count += 1
			return &implI{1}
		})
		call := m.EchoHandler(func(_ echo.Context, _ I) error {
			return nil
		})
		require.NoError(t, call(ctx))
		require.NoError(t, call(ctx))
		require.Equal(t, 1, count)
		m.Reset()

		require.NoError(t, call(ctx))
		require.Equal(t, 2, count)
	})

	t.Run("err - build of type failed", func(t *testing.T) {
		m := magnet.New()
		m.Register(func() (A, error) {
			return A{}, errors.New("failed to build A")
		})

		c := m.NewCaller(func(A) error {
			return nil
		})

		_, err := c.Call()
		require.Error(t, err, "failed to build A")
	})

	t.Run("panic - types cannot be built", func(t *testing.T) {
		type B struct{}
		m := magnet.New()
		m.Register(func() (A, error) {
			return A{}, nil
		})

		expectedPanic := magnetErrors.NewCannotBeBuiltErr(reflect.TypeOf(B{}))
		require.PanicsWithValue(t, expectedPanic, func() {
			m.NewCaller(func(B) error {
				return nil
			})
		})
	})

	t.Run("panic - cycle", func(t *testing.T) {
		m := magnet.New()

		m.Register(func(A) A { return A{} })
		require.Panics(t, func() {
			m.NewCaller(func(A) error { return nil })
		}, "cycle found!")
	})

	t.Run("panic - large cycle", func(t *testing.T) {
		m := magnet.New()

		type A struct{}
		type B struct{}
		type C struct{}

		msg := recoverPanic(func() {
			m.Register(func(A) B { return B{} })
			m.Register(func(B) C { return C{} })
			m.Register(func(C) A { return A{} })
			m.NewCaller(func(A) error { return nil })
		})
		checkCycleError(t, []string{"A", "B", "C"}, msg)
	})

	t.Run("panic - complex cycle", func(t *testing.T) {
		m := magnet.New()

		type A struct{}
		type B struct{}
		type C struct{}

		msg := recoverPanic(func() {
			m.Register(func(A, C) B { return B{} })
			m.Register(func(A, B) C { return C{} })
			m.Register(func(B, C) A { return A{} })
			m.NewCaller(func(A) error { return nil })
		})
		checkCycleError(t, []string{"A", "B"}, msg)
	})

	t.Run("panic - cycle through parent", func(t *testing.T) {
		m := magnet.New()

		type A struct{}
		type B struct{}

		msg := recoverPanic(func() {
			m.Register(func() B { return B{} })
			m.Register(func(B) A { return A{} })
			c := m.NewChild()
			c.Register(func(A) B { return B{} })
			c.NewCaller(func(A) error { return nil })
		})
		checkCycleError(t, []string{"A", "B"}, msg)
	})
}
