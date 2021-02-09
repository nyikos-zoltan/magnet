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

type AnonDerivedStruct struct {
	magnet.Derived
	I
}

func Test_Magnet(t *testing.T) {
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

		require.Panics(t, func() {
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
			m.NewCaller(func(A) error { return nil })
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
			m.NewCaller(func(A) error { return nil })
		})
	})
}
