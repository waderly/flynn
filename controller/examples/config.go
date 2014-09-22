package main

import (
	"fmt"
	"os"
)

type config struct {
	controllerDomain string
	controllerKey    string
}

func loadConfigFromEnv() (c *config, err error) {
	c = &config{}
	c.controllerDomain = os.Getenv("CONTROLLER_DOMAIN")
	if c.controllerDomain == "" {
		err = fmt.Errorf("CONTROLLER_DOMAIN is required")
	}
	c.controllerKey = os.Getenv("CONTROLLER_KEY")
	if c.controllerKey == "" {
		err = fmt.Errorf("CONTROLLER_KEY is required")
	}
	return c, err
}
