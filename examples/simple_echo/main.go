package main

import (
	"github.com/labstack/echo/v4"
	"github.com/nyikos-zoltan/magnet"
)

type Config struct {
	ConfigValue string
}

type RootArgs struct {
	Arg string `json:"arg"`
}

type RootResponse struct {
	ConfigValue string
	Arg         string
}

func RootHandler(ctx echo.Context, c *Config, args RootArgs) error {
	return ctx.JSON(200, RootResponse{
		Arg:         args.Arg,
		ConfigValue: c.ConfigValue,
	})
}

func main() {
	m := magnet.New()
	m.RegisterMany(
		func() *Config {
			return &Config{"some_value"}
		},
		func(e echo.Context) (r RootArgs, err error) {
			err = e.Bind(&r)
			return
		},
	)

	e := echo.New()
	e.POST("/", m.EchoHandler(RootHandler))
	if err := e.Start(":8081"); err != nil {
		panic(err.Error())
	}
}
