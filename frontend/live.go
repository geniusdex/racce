package frontend

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/geniusdex/racce/accserver"
)

type livePage struct {
	Server *accserver.Server
}

func (f *frontend) liveHandler(w http.ResponseWriter, r *http.Request) {
	page := &livePage{
		Server: f.server,
	}

	f.executeTemplate(w, r, "live.html", page)
}

func (f *frontend) liveWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := newWriteOnlyWebSocket(w, r)
	if err != nil {
		log.Panicf("Failed to create websocket: %v", err)
	}

	go f.sendLiveStateUpdates(ws)
}

func writeMessageToWebSocket(ws *writeOnlyWebSocket, msgType string, data interface{}) error {
	jsonMsg, err := json.Marshal(map[string]interface{}{"type": msgType, "data": data})
	if err != nil {
		log.Printf("cannot marshal message as JSON: %v", err)
	}

	return ws.WriteTextMessage(jsonMsg)
}

func (f *frontend) sendLiveStateUpdates(ws *writeOnlyWebSocket) {
	log.Printf("Sending live state updates on websocket connection %v", ws.Name())

	events := f.server.LiveState.NewEventChannels()
	defer func() {
		ws.Close()
		events.Flush()
	}()

	for {
		select {
		case state, ok := <-events.ServerState:
			if !ok {
				return
			}
			writeMessageToWebSocket(ws, "serverState", state)

		case nrClients, ok := <-events.NrClients:
			if !ok {
				return
			}
			writeMessageToWebSocket(ws, "nrClients", nrClients)

		case track, ok := <-events.Track:
			if !ok {
				return
			}
			writeMessageToWebSocket(ws, "track", track)
		}
	}
}
