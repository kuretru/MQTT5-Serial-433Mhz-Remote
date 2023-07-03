package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"mqtt5-serial-433mhz-remote/internal"
	"mqtt5-serial-433mhz-remote/internal/api"
	"mqtt5-serial-433mhz-remote/internal/mqtt"
	"mqtt5-serial-433mhz-remote/internal/serialport"
)

func parseArgument() string {
	var configFile string
	flag.StringVar(&configFile, "config", "/etc/mqtt5-serial-433mhz-remote.yaml", "The path to the configuration file")
	flag.Parse()

	_, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		log.Fatalf("[Core] Config file not exist: %s\n", configFile)
	}

	return configFile
}

func main() {
	configFile := parseArgument()
	appConfig := internal.LoadConfig(configFile)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	commandChannel, stateChannel := make(chan string), make(chan string)
	wg.Add(3)
	go serialport.InstallSerial(ctx, wg.Done, appConfig, commandChannel, stateChannel)
	go mqtt.InstallMQTTClient(ctx, wg.Done, appConfig, commandChannel, stateChannel)
	go api.InstallWebAPI(ctx, wg.Done, appConfig, commandChannel)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	log.Println("[Core] Signal caught - exiting")
	cancel()
	wg.Wait()
}
