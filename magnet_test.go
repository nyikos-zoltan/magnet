package magnet_test

import (
	"errors"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nyikos-zoltan/magnet"
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

func Test_Magnet(t *testing.T) {
	ctx := echo.New().NewContext(nil, nil)
	t.Run("ok - simple", func(t *testing.T) {
		m := magnet.New()
		m.Register(func() A {
			return A{}
		})

		ctx := echo.New().NewContext(nil, nil)

		var injA *A
		require.NoError(t, m.EchoHandler(func(a A) error {
			injA = &a
			return nil
		})(ctx))
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

	t.Run("ok - ctx", func(t *testing.T) {
		m := magnet.New()

		ctx := echo.New().NewContext(nil, nil)

		var injCtx echo.Context
		require.NoError(t, m.EchoHandler(func(c echo.Context) error {
			injCtx = c
			return nil
		})(ctx))
		require.NotNil(t, injCtx)
	})
	t.Run("ok - only values that need ctx are recreated", func(t *testing.T) {
		m := magnet.New()
		type B struct{ int }
		aCount := 0
		m.Register(func() *A {
			aCount++
			return &A{}
		})

		bCount := 0
		m.Register(func(_ echo.Context) *B {
			bCount++
			return &B{0}
		})

		var injA1, injA2 *A
		var injB1, injB2 *B
		require.NoError(t, m.EchoHandler(func(a *A, b *B) error {
			injA1 = a
			injB1 = b
			return nil
		})(ctx))
		require.NoError(t, m.EchoHandler(func(a *A, b *B) error {
			injA2 = a
			injB2 = b
			return nil
		})(ctx))
		require.Same(t, injA1, injA2)
		require.False(t, injB1 == injB2) // no opposite of Same
		require.EqualValues(t, 1, aCount)
		require.EqualValues(t, 2, bCount)
	})

	t.Run("err - build of type failed", func(t *testing.T) {
		m := magnet.New()
		m.Register(func() (A, error) {
			return A{}, errors.New("failed to build A")
		})

		echoh := m.EchoHandler(func(A) error {
			return nil
		})

		require.Error(t, echoh(ctx), "failed to build A")
	})

	t.Run("err - handler failed", func(t *testing.T) {
		m := magnet.New()
		m.Register(func() (A, error) {
			return A{}, nil
		})

		echoh := m.EchoHandler(func(A) error {
			return errors.New("handler failed")
		})

		require.Error(t, echoh(ctx), "handler failed")
	})

	t.Run("panic - types cannot be built", func(t *testing.T) {
		type B struct{}
		m := magnet.New()
		m.Register(func() (A, error) {
			return A{}, nil
		})

		require.Panics(t, func() {
			m.EchoHandler(func(B) error {
				return nil
			})
		})
	})

	t.Run("panic - handler fn invalid", func(t *testing.T) {
		m := magnet.New()

		require.Panics(t, func() {
			m.EchoHandler(func(A) {
			})
		})
	})

	t.Run("panic - cycle", func(t *testing.T) {
		m := magnet.New()

		require.Panics(t, func() {
			m.Register(func(A) A { return A{} })
			_ = m.EchoHandler(func(A) error { return nil })(ctx)
		})
	})

	t.Run("panic - large cycle", func(t *testing.T) {
		m := magnet.New()

		type A struct{}
		type B struct{}
		type C struct{}

		require.Panics(t, func() {
			m.Register(func(A) B { return B{} })
			m.Register(func(B) C { return C{} })
			m.Register(func(C) A { return A{} })
			_ = m.EchoHandler(func(A) error { return nil })(ctx)
		})
	})

	t.Run("panic - complex cycle", func(t *testing.T) {
		m := magnet.New()

		type A struct{}
		type B struct{}
		type C struct{}

		require.Panics(t, func() {
			m.Register(func(A, B, C) B { return B{} })
			m.Register(func(A, B, C) C { return C{} })
			m.Register(func(A, B, C) A { return A{} })
			_ = m.EchoHandler(func(A) error { return nil })(ctx)
		})
	})
}
