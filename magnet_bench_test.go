package magnet_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/loov/hrtime/hrtesting"
	"github.com/nyikos-zoltan/magnet"
)

type SomeFactory struct {
	n int
}
type OtherFactory struct {
	n int
}
type OtherFactory2 struct {
	n int
}

type reply struct {
	Sf int
	Of int
}

type requestData struct {
	Stuff stuff `json:"some-stuff"`
}
type stuff struct {
	Value int `json:"value"`
}

func Benchmark_Echo(b *testing.B) {
	rvb := echo.New()
	rd := strings.NewReader(`{"some-stuff": {"value": 1}}`)
	req := httptest.NewRequest("POST", "/", rd)

	staticCounter := 0
	dynCounter := 0
	sf := SomeFactory{n: staticCounter}

	rvb.POST("/", func(ctx echo.Context) error {
		dynCounter++
		of := OtherFactory{n: dynCounter}
		var params requestData
		if err := ctx.Bind(&params); err != nil {
			return err
		}
		return ctx.JSON(http.StatusOK, reply{Sf: sf.n, Of: of.n})
	})

	rec := httptest.NewRecorder()
	rvb.ServeHTTP(rec, req)

	bench := hrtesting.NewBenchmark(b)
	defer bench.Report()

	for bench.Next() {
		rec.Body.Truncate(0)
		rvb.ServeHTTP(rec, req)
	}
}

func Benchmark_Magnet1(b *testing.B) {
	m := magnet.New()
	rd := strings.NewReader(`{"some-stuff": {"value": 1}}`)
	req := httptest.NewRequest("POST", "/", rd)

	staticCounter := 0
	m.Register(func() SomeFactory {
		staticCounter++
		return SomeFactory{n: staticCounter}
	})
	dynCounter := 0
	m.Register(func(echo.Context) OtherFactory {
		dynCounter++
		return OtherFactory{n: dynCounter}
	})

	e := echo.New()

	e.POST("/", m.EchoHandler(func(sf SomeFactory, of OtherFactory, ctx echo.Context) error {
		var params requestData
		if err := ctx.Bind(&params); err != nil {
			return err
		}
		return ctx.JSON(http.StatusOK, reply{Sf: sf.n, Of: of.n})
	}))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	bench := hrtesting.NewBenchmark(b)
	defer bench.Report()

	for bench.Next() {
		rec.Body.Truncate(0)
		e.ServeHTTP(rec, req)
	}
}

func Benchmark_Magnet2(b *testing.B) {
	a := magnet.New()
	rd := strings.NewReader(`{"some-stuff": {"value": 1}}`)
	req := httptest.NewRequest("POST", "/", rd)

	staticCounter := 0
	a.Register(func() SomeFactory {
		staticCounter++
		return SomeFactory{n: staticCounter}
	})
	a.Register(func(SomeFactory) OtherFactory2 {
		return OtherFactory2{n: 1}
	})
	dynCounter := 0
	a.Register(func(echo.Context) OtherFactory {
		dynCounter++
		return OtherFactory{n: dynCounter}
	})

	e := echo.New()
	e.POST("/", a.EchoHandler(func(sf SomeFactory, of OtherFactory, of2 OtherFactory2, ctx echo.Context) error {
		var params requestData
		if err := ctx.Bind(&params); err != nil {
			return err
		}
		return ctx.JSON(http.StatusOK, reply{Sf: sf.n, Of: of.n})
	}))

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	bench := hrtesting.NewBenchmark(b)
	defer bench.Report()

	for bench.Next() {
		rec.Body.Truncate(0)
		e.ServeHTTP(rec, req)
	}
}
