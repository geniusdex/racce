package frontend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/geniusdex/racce/accserver"
)

type adminServerLogPage struct {
	Server *accserver.Server
}

func (a *admin) serverLogHandler(w http.ResponseWriter, r *http.Request) {
	a.executeTemplate(w, r, "admin-server-log.html", &adminServerLogPage{a.server})
}

func (a *admin) serverLogWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	serverInstance := a.server.Instance
	if serverInstance == nil {
		log.Panicf("No running server instance")
	}

	ws, err := newWriteOnlyWebSocket(w, r)
	if err != nil {
		log.Panicf("Failed to create websocket: %v", err)
	}

	go writeServerLogToWebSocket(serverInstance.NewLogChannel(), ws)
}

func writeLogMessageToWebSocket(msg accserver.LogMessage, ws *writeOnlyWebSocket) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("cannot marshal message as JSON: %w", err)
	}

	return ws.WriteTextMessage(jsonMsg)
}

func writeServerLogToWebSocket(logChannel <-chan accserver.LogMessage, ws *writeOnlyWebSocket) {
	log.Printf("Sending server log on websocket connection %v", ws.Name())

	for {
		msg, ok := <-logChannel
		if !ok {
			break
		}

		writeLogMessageToWebSocket(msg, ws)
	}

	ws.Close()
}
