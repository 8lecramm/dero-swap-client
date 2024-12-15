package main

import (
	"log"
	"os"

	"swap-client/cfg"
	"swap-client/coin"

	"github.com/robfig/cron/v3"
)

func Init() {

	// create all swap directories
	os.MkdirAll("swaps/active", 0755)
	os.MkdirAll("swaps/expired", 0755)
	os.MkdirAll("swaps/done", 0755)
}

func main() {

	Init()
	cfg.LoadConfig()
	if !cfg.CheckConfig() {
		log.Println("Configuration error. Please check config file!")
		os.Exit(1)
	}

	cfg.LoadWallets()
	coin.Locked.LoadLockedBalance()

	c := cron.New()
	c.AddFunc("@every 1m", UpdateMarkets)
	c.AddFunc("@every 2m", Delay.CheckBackoff)
	c.Start()

	go Swap_Controller()
	StartClient(cfg.Server_URL)
}
