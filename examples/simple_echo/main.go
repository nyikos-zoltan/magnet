package main

import (
	"github.com/labstack/echo/v4"
	"github.com/nyikos-zoltan/magnet"
)

type Config struct {
	ConfigValue string
}

func RootHandler(ctx echo.Context, c *Config) error {
	return ctx.JSON(200, c)
}

func main() {
	m := magnet.New()
	m.RegisterMany(
		func() *Config {
			return &Config{"some_value"}
		},
	)

	e := echo.New()
	e.GET("/", m.EchoHandler(RootHandler))
	if err := e.Start(":8081"); err != nil {
		panic(err.Error())
	}
}
