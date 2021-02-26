package main

import (
	"fmt"

	"github.com/nyikos-zoltan/magnet"
)

type Config struct {
	SomeValue string
}

func main() {
	m := magnet.New()
	m.Register(func() *Config {
		return &Config{SomeValue: "test config value"}
	})
	var cfg *Config
	if err := m.Retrieve(&cfg); err != nil {
		panic(err)
	}
	fmt.Println(cfg.SomeValue)
}
