package main

import (
	"log"

	"github.com/nyikos-zoltan/magnet"
)

type Config struct {
	ConfigValue string
}

func main() {
	m := magnet.New()
	m.RegisterMany(
		func() *Config {
			return &Config{"some_value"}
		},
	)

	caller := m.NewCaller(func(c *Config) string { return c.ConfigValue })
	rv, err := caller.Call()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(rv[0].String())
}
