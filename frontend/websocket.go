package frontend

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

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
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type webSocketMessage struct {
	messageType int
	data        []byte
	err         chan error
}

// writeOnlyWebSocket is a websocket which is only used to write to the client
type writeOnlyWebSocket struct {
	connection *websocket.Conn
	closed     bool
	messages   chan *webSocketMessage
	close      chan bool
}

// newWriteOnlyWebSocket creates a new websocket by upgrading a HTTP connection
func newWriteOnlyWebSocket(w http.ResponseWriter, r *http.Request) (*writeOnlyWebSocket, error) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot upgrade connection to websocket: %w", err)
	}

	ws := &writeOnlyWebSocket{
		connection: connection,
		closed:     false,
		messages:   make(chan *webSocketMessage),
		close:      make(chan bool),
	}

	log.Printf("New websocket connection %v", ws.Name())

	go ws.writeMessages()
	go ws.readMessages()

	return ws, nil
}

func (ws *writeOnlyWebSocket) Name() string {
	return fmt.Sprintf("to %v", ws.connection.RemoteAddr().String())
}

// writeMessage writes a message of a specific type
func (ws *writeOnlyWebSocket) writeMessage(messageType int, data []byte) error {
	if ws.closed {
		return fmt.Errorf("cannot send message: websocket already closed")
	}

	err := make(chan error)
	ws.messages <- &webSocketMessage{messageType, data, err}
	return <-err
}

// WriteTextMessage writes a UTF-8 encoded text message to the websocket
func (ws *writeOnlyWebSocket) WriteTextMessage(data []byte) error {
	return ws.writeMessage(websocket.TextMessage, data)
}

// Close sends a close message to the client, which will then close the socket
func (ws *writeOnlyWebSocket) Close() {
	ws.writeMessage(websocket.CloseMessage, []byte{})
	ws.close <- true
}

// closeConnection closes the actual network connection without sending a close message
func (ws *writeOnlyWebSocket) closeConnection() {
	if !ws.closed {
		log.Printf("Closing websocket connection %v", ws.Name())
		if err := ws.connection.Close(); err != nil {
			log.Printf("Websocket %v close error: %v", ws.Name(), err)
		}
		close(ws.messages)
		ws.closed = true
	}
}

// writeMessageOnSocket writes a message on the websocket itself
func (ws *writeOnlyWebSocket) writeMessageOnSocket(messageType int, data []byte) error {
	ws.connection.SetWriteDeadline(time.Now().Add(writeWait))
	err := ws.connection.WriteMessage(messageType, data)

	if !isExpectedWebSocketCloseError(err) {
		log.Printf("Websocket %v write error: %v", ws.Name(), err)
	}

	return err
}

// writeMessages sends any message coming on a write channel over the websocket
func (ws *writeOnlyWebSocket) writeMessages() {
	pingTicker := time.NewTicker(pingPeriod)

loop:
	for {
		var err error = nil

		select {
		case msg, ok := <-ws.messages:
			if !ok {
				break loop
			}
			err = ws.writeMessageOnSocket(msg.messageType, msg.data)
			msg.err <- err

		case <-ws.close:
			break loop

		case <-pingTicker.C:
			err = ws.writeMessageOnSocket(websocket.PingMessage, nil)
		}

		if err != nil {
			break loop
		}
	}

	pingTicker.Stop()
	ws.closeConnection()
}

// readMessages monitors the incoming channel so pong messages are handled properly
func (ws *writeOnlyWebSocket) readMessages() {
	ws.connection.SetReadLimit(maxMessageSize)
	ws.connection.SetReadDeadline(time.Now().Add(pongWait))
	ws.connection.SetPongHandler(func(string) error {
		ws.connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := ws.connection.ReadMessage()
		if err != nil {
			if !isExpectedWebSocketCloseError(err) {
				log.Printf("Websocket %v read error: %v", ws.Name(), err)
			}
			break
		}
	}

	ws.close <- true
}

// isExpectedWebSocketCloseError checks if a given value is an error indicate a websocket is closed in some expected way
func isExpectedWebSocketCloseError(v interface{}) bool {
	if err, ok := v.(error); ok {
		err = errors.Unwrap(err)
		return !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure)
	}
	return true
}
