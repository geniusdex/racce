package frontend

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/geniusdex/racce/accserver"
	"github.com/gorilla/websocket"
)

type adminServerLogPage struct {
	Server *accserver.Server
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 30 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func (a *admin) serverLogHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, r, "admin-server-log.html", &adminServerLogPage{a.server})
}

func connStr(connection *websocket.Conn) string {
	return fmt.Sprintf("to %v", connection.RemoteAddr().String())
}

func (a *admin) serverLogWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	serverInstance := a.server.Instance
	if serverInstance == nil {
		log.Panicf("No running server instance")
	}

	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panicf("Cannot upgrade connection to websocket: %v", err)
	}

	log.Printf("New server log websocket connection %v", connStr(connection))

	go serverLogToWebSocket(connection, serverInstance.NewLogChannel())
	go serverLogWebSocketReadMonitoring(connection)
}

func isExpectedWebSocketCloseError(v interface{}) bool {
	if err, ok := v.(error); ok {
		err = errors.Unwrap(err)
		return !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure)
	}
	return false
}

func writeLogMessageToWebSocket(msg accserver.LogMessage, connection *websocket.Conn) {
	writer, err := connection.NextWriter(websocket.TextMessage)
	if err != nil {
		panic(fmt.Errorf("cannot create new writer for websocket: %w", err))
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		panic(fmt.Errorf("cannot marshal message as JSON: %w", err))
	}
	writer.Write(jsonMsg)
	writer.Write(newline)
	if err := writer.Close(); err != nil {
		panic(fmt.Errorf("cannot write log message to websocket: %w", err))
	}
}

func serverLogToWebSocket(connection *websocket.Conn, logChannel <-chan accserver.LogMessage) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		if r := recover(); r != nil && !isExpectedWebSocketCloseError(r) {
			log.Printf("Websocket write error: %v", r)
		}

		log.Printf("Closing write channel for websocket connection %v", connStr(connection))

		pingTicker.Stop()
		connection.Close()
	}()
	for {
		select {
		case msg, ok := <-logChannel:
			connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writeLogMessageToWebSocket(msg, connection)

		case <-pingTicker.C:
			connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				panic(fmt.Errorf("cannot write ping message to websocket: %w", err))
			}
		}
	}
}

// serverLogWebSocketReadMonitoring monitors the incoming channel so pong messages are handled properly
func serverLogWebSocketReadMonitoring(connection *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil && !isExpectedWebSocketCloseError(r) {
			log.Printf("Websocket read error: %v", r)
		}

		log.Printf("Closing read channel for websocket connection %v", connStr(connection))

		connection.Close()
	}()

	connection.SetReadLimit(maxMessageSize)
	connection.SetReadDeadline(time.Now().Add(pongWait))
	connection.SetPongHandler(func(string) error {
		connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := connection.ReadMessage()
		if err != nil {
			panic(fmt.Errorf("cannot read messages from websocket: %w", err))
		}
	}
}
