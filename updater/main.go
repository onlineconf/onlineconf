package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/onlineconf/onlineconf/updater/v3/updater"
)

var configFile = flag.String("config", "/usr/local/etc/onlineconf.yaml", "config file")
var once = flag.Bool("once", false, "fetch configuration once and exit")

func main() {
	flag.Parse()
log.Info().Str("config", *configFile).Msg("bla")
	config := readConfigFile(*configFile)
	u := updater.NewUpdater(*config)

	if *once {
		if u.Update() != nil {
			os.Exit(1)
		}
		return
	}

	log.Info().Msg("onlineconf-updater started")
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigC)
	go func() {
		sig := <-sigC
		log.Info().Str("signal", sig.String()).Msg("signal received, terminating")
		signal.Stop(sigC)
		u.Stop()
	}()

	u.Update()
	u.Run()
	log.Info().Msg("onlineconf-updater stopped")
}
