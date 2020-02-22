package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/barelyuseless/simple-service-registry/internal"
)

func initViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Err(err).Msg("viper couldn't read in a config file - defaults/ENV will be used")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func initServiceStore() (internal.ServiceStore, error) {
	var store internal.ServiceStore
	var err error

	switch viper.GetString("servicestore.type") {
	case "boltdb":
		path := viper.GetString("servicestore.path")
		if path == "" {
			path = "./servicestore.db"
			log.Warn().Msgf("Using boltdb store without an explicitly set path; defaulting to %s", path)
		}

		store, err = internal.NewBoltDBServiceStore(path)
		if err != nil {
			return nil, err
		}
	default:
		store = internal.NewMapServiceStore()
	}

	if viper.GetBool("servicestore.polling.use") {
		intervalString := viper.GetString("servicestore.polling.interval")
		var interval time.Duration
		if intervalString == "" {
			interval = 60 * time.Second
			log.Warn().Msgf("Using polling servicestore without an explicitly set interval; defaulting to %v", interval)
		} else {
			interval, err = time.ParseDuration(intervalString)
			if err != nil {
				return nil, err
			}
		}
		store, err = internal.NewPollingServiceStore(store, interval)
	}

	return store, nil
}

func main() {
	initViper()

	serviceStore, err := initServiceStore()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Info().Msgf("Initialised servicestore of type %s", serviceStore)

	address := viper.GetString("address")
	if address == "" {
		address = "0.0.0.0:8080"
		log.Warn().Msgf("No bind address specified; defaulting to %s", address)
	}

	server := internal.StartAPI(address, serviceStore)

	log.Info().Msgf("Initialised service api on %s", address)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	log.Info().Msg("Recieved interrupt, shutting down...")

	// Shutdown API server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Shutdown(ctx)
	if err != nil {
		log.Err(err).Msg("Error while shutting API server")
	}

	// Close serviceStore
	err = serviceStore.Close()
	if err != nil {
		log.Err(err).Msg("Error while shutting down servicestore")
	}

	log.Info().Msg("shutdown complete, terminating.")
}
