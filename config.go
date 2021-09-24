package main

import (
	"os"

	"github.com/apex/log"
)

// TODO
func loadConfig(configFile string) (*config, error) {
	return &config{}, nil
}

type config struct {
	Title string
	Icon  string
	Pwd   string

	Items []string
}

func process(c *config) {
	if c.Pwd != "" {
		if err := os.Chdir(c.Pwd); err != nil {
			log.Errorf("ERR: %s", err)
			return
		}
	}
}
