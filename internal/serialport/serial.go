package serialport

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	"mqtt5-serial-433mhz-remote/internal"

	"go.bug.st/serial"
)

var (
	serialMode = &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
)

func InstallSerial(ctx context.Context, done func(), config *internal.AppConfig, commandChan chan string, stateChan chan string) {
	defer done()

	serialPort, err := serial.Open(config.Serial.Port, serialMode)
	if err != nil {
		if config.Debug {
			log.Printf("[Serial] Port %s not found, but debug mode was enabled, entering fake serial mode\n", config.Serial.Port)
			serialPort = nil
		} else {
			log.Fatalf("[Serial] Open port %s failed, %s\n", config.Serial.Port, err)
		}
	}
	defer func(serialPort serial.Port) {
		if serialPort != nil {
			log.Println("[Serial] Exiting")
			if err := serialPort.Close(); err != nil {
				log.Fatalf("[Serial] Close port %s failed, %s\n", config.Serial.Port, err)
			}
		}
	}(serialPort)

	stopChan := make(chan string, 1)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case command := <-commandChan:
			serialSend(serialPort, config, command, stateChan, stopChan)
		}
	}
}

func serialSend(serialPort serial.Port, config *internal.AppConfig, command string, stateChan chan string, stopChan chan string) {
	log.Printf("[Serial] Get command \"%s\"\n", command)

	select {
	case <-stopChan:
	default:
	}

	commandData := ""
	switch command {
	case "OPEN":
		commandData = config.Serial.OpenCommand
		stateChan <- "opening"
		go func() {
			select {
			case <-time.After(time.Duration(config.Serial.OpenTime) * time.Second):
				stateChan <- "open"
			case <-stopChan:
				stateChan <- "open"
				return
			}
		}()
	case "CLOSE":
		commandData = config.Serial.CloseCommand
		stateChan <- "closing"
		go func() {
			select {
			case <-time.After(time.Duration(config.Serial.CloseTime) * time.Second):
				stateChan <- "closed"
			case <-stopChan:
				stateChan <- "open"
				return
			}
		}()
	case "STOP":
		commandData = config.Serial.StopCommand
		stateChan <- "stopped"
		stopChan <- "stopped"
	}

	if commandData == "" {
		log.Printf("[Serial] Unknown command \"%s\"\n", command)
		return
	}
	data, err := hex.DecodeString(commandData)
	if err != nil {
		panic(err)
	}

	if serialPort == nil {
		log.Printf("[Serial] Fake serial, sending \"%s\"\n", commandData)
	} else {
		_, err = serialPort.Write(data)
		if err != nil {
			log.Fatalf("[Serial] Send data failed, %s\n", err)
		}
	}
}
