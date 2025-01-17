package magnet_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nyikos-zoltan/magnet"
	"github.com/nyikos-zoltan/magnet/gorm_v1"
	magnetErrs "github.com/nyikos-zoltan/magnet/internal/errors"
	"github.com/nyikos-zoltan/magnet/transaction"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EchoTestSuite struct {
	suite.Suite
	ctx echo.Context
	m   *magnet.Magnet
}

func (s *EchoTestSuite) SetupTest() {
	s.ctx = echo.New().NewContext(nil, nil)
	s.m = magnet.New()
}

func (s *EchoTestSuite) TestOkCtx() {
	var injCtx echo.Context
	require.NoError(s.T(), s.m.EchoHandler(func(c echo.Context) error {
		injCtx = c
		return nil
	})(s.ctx))
	require.NotNil(s.T(), injCtx)
}

func (s *EchoTestSuite) TestOkRecreateCtxOnly() {
	type B struct{ int }
	aCount := 0
	s.m.Register(func() *A {
		aCount++
		return &A{}
	})

	bCount := 0
	s.m.Register(func(_ echo.Context) *B {
		bCount++
		return &B{0}
	})

	var injA1, injA2 *A
	var injB1, injB2 *B
	require.NoError(s.T(), s.m.EchoHandler(func(a *A, b *B) error {
		injA1 = a
		injB1 = b
		return nil
	})(s.ctx))
	require.NoError(s.T(), s.m.EchoHandler(func(a *A, b *B) error {
		injA2 = a
		injB2 = b
		return nil
	})(s.ctx))
	require.Same(s.T(), injA1, injA2)
	require.False(s.T(), injB1 == injB2) // no opposite of Same
	require.EqualValues(s.T(), 1, aCount)
	require.EqualValues(s.T(), 2, bCount)
}

func (s *EchoTestSuite) TestErrHandlerFailed() {
	s.m.Register(func() (A, error) {
		return A{}, nil
	})

	echoh := s.m.EchoHandler(func(A) error {
		return errors.New("handler failed")
	})

	require.Error(s.T(), echoh(s.ctx), "handler failed")
}

func (s *EchoTestSuite) TestPanicTypesCannotBeBuilt() {
	type B struct{}
	s.m.Register(func() (A, error) {
		return A{}, nil
	})

	require.Panics(s.T(), func() {
		s.m.EchoHandler(func(B) error {
			return nil
		})
	})
}

func (s *EchoTestSuite) TestPanicHandlerFnInvalid() {
	s.m.Register(func() (A, error) {
		return A{}, nil
	})

	require.Panics(s.T(), func() {
		s.m.EchoHandler(func(A) {
		})
	})
}
func (s *EchoTestSuite) TestPanicHandlerCantBeBuilt() {
	gorm_v1.Plugin(s.m)
	type txType = func(func(transaction.Tx, A) error) error

	expectedValue := magnetErrs.NewCannotBeBuiltErr(reflect.TypeOf(A{}))
	require.PanicsWithValue(s.T(), expectedValue, func() {
		s.m.EchoHandler(func(a txType) error {
			return nil
		})
	})
}

func (s *EchoTestSuite) TestOkErrorHandler() {
	req := s.Require()
	type a struct{ int }
	s.m.Value(a{1})
	called := false
	handler := s.m.EchoErrorHandler(func(e error, ctx echo.Context, a a) {
		req.EqualError(e, "some error")
		req.Equal(s.ctx, ctx)
		req.Equal(1, a.int)
		called = true
	})
	handler(errors.New("some error"), s.ctx)
	req.True(called)
}

func (s *EchoTestSuite) TestPanicErrorHandlerInvalid() {
	req := s.Require()
	req.PanicsWithValue("EchoErrorHandler methods need to take at least an error and the echo.Context", func() {
		s.m.EchoErrorHandler(func() {})
	})
}

func TestEchoSuite(t *testing.T) {
	suite.Run(t, new(EchoTestSuite))
}
