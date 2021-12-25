package main

import (
	"log"
	"mkm-watchdog/config"
	"mkm-watchdog/coordinator"
	"time"
)

func main() {
	log.Println("Starting mkm-watchdog")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error while loading config: %v", err)
	}

	tpl, err := cfg.LoadTemplate()
	if err != nil {
		log.Fatalf("Could not parse message template: %v\n", err)
	}

	sleepPeriod := time.Duration(cfg.Delay) * time.Second

	c := coordinator.NewCoordinator(cfg.Searches, sleepPeriod, tpl)
	c.Start()
}
