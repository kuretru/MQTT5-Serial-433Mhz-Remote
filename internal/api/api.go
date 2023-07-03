package api

import (
	"context"
	"io"
	"log"
	"net/http"

	"mqtt5-serial-433mhz-remote/internal"
)

var (
	sendCommand chan string
)

func InstallWebAPI(ctx context.Context, done func(), config *internal.AppConfig, commandChannel chan string) {
	defer done()

	sendCommand = commandChannel

	httpServer := http.Server{Addr: config.API.Listen}
	http.HandleFunc("/open_door", openDoor)
	http.HandleFunc("/close_door", closeDoor)
	http.HandleFunc("/stop_door", stopDoor)

	go func(server *http.Server) {
		log.Printf("[Web API] Server start at %s\n", config.API.Listen)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Web API] Failed start Web API at %s, %s\n", config.API.Listen, err)
		}
	}(&httpServer)

	<-ctx.Done()
	log.Println("[Web API] Exiting")
	err := httpServer.Close()
	if err != nil {
		log.Printf("[Web API] Close server failed %s\n", err)
	}
}

func openDoor(w http.ResponseWriter, req *http.Request) {
	log.Println("[Web API]: open door")
	sendCommand <- "OPEN"
	safeWriteResponse(w, "opened")
}

func closeDoor(w http.ResponseWriter, req *http.Request) {
	log.Println("[Web API]: close door")
	sendCommand <- "CLOSE"
	safeWriteResponse(w, "closed")
}

func stopDoor(w http.ResponseWriter, req *http.Request) {
	log.Println("[Web API]: stop door")
	sendCommand <- "STOP"
	safeWriteResponse(w, "stopped")
}

func safeWriteResponse(w http.ResponseWriter, data string) {
	_, err := io.WriteString(w, data)
	if err != nil {
		log.Println("Web API: Send response error: ", err)
	}
}
