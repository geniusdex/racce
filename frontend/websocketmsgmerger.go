package frontend

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	// wmmMaxMessageSize is the maximum size of a message; a single message is never split
	wmmMaxMessageSize = 64 * 1024
	// mergePeriod is the time period in which messages are merged
	mergePeriod = 25 * time.Millisecond
)

// webSocketMessageMerger is a wrapper around writeOnlyWebSocket that merges all messages sent within a short interval
type webSocketMessageMerger struct {
	ws        *writeOnlyWebSocket
	messages  chan []byte
	closed    bool
	lastError error
}

// newWebSocketMessageMerger creates a new writeOnlyWebSocket and message merge wrapper around it
func newWebSocketMessageMerger(w http.ResponseWriter, r *http.Request) (*webSocketMessageMerger, error) {
	ws, err := newWriteOnlyWebSocket(w, r)

	wmm := &webSocketMessageMerger{
		ws:        ws,
		messages:  make(chan []byte),
		closed:    false,
		lastError: nil,
	}

	go wmm.mergeMessages()

	return wmm, err
}

// Name returns a printable identifier for this websocket
func (wmm *webSocketMessageMerger) Name() string {
	return wmm.ws.Name()
}

// WriteTextMessage writes a UTF-8 encoded text message to the websocket
func (wmm *webSocketMessageMerger) WriteTextMessage(data []byte) error {
	if wmm.closed {
		return fmt.Errorf("websocket %v already closed", wmm.Name())
	} else if wmm.lastError != nil {
		return wmm.lastError
	} else {
		wmm.messages <- data
		return nil
	}
}

// Close sends a close message to the client, which will then close the socket
func (wmm *webSocketMessageMerger) Close() {
	if !wmm.closed {
		close(wmm.messages)
		wmm.ws.Close()
		wmm.closed = true
	}
}

func (wmm *webSocketMessageMerger) sendBuffer(buffer []byte) error {
	err := wmm.ws.WriteTextMessage(buffer)
	if err != nil {
		log.Printf("Stopping message merging for websocket %v (%v)", wmm.Name(), err)
		wmm.lastError = err
	}
	return err
}

// mergeMessages reads the written messages and merges them periodically
func (wmm *webSocketMessageMerger) mergeMessages() {
	var bufferTimer <-chan time.Time = nil
	buffer := make([]byte, 0, wmmMaxMessageSize)

loop:
	for {
		select {
		case msg, ok := <-wmm.messages:
			if !ok {
				break loop
			}
			if len(buffer)+len(msg) >= wmmMaxMessageSize {
				if err := wmm.sendBuffer(buffer); err != nil {
					break loop
				}
				buffer = buffer[:0]
			}
			if len(buffer) > 0 {
				buffer = append(buffer, '\n')
			}
			buffer = append(buffer, msg...)
			if bufferTimer == nil {
				bufferTimer = time.After(mergePeriod)
			}

		case <-bufferTimer:
			if err := wmm.sendBuffer(buffer); err != nil {
				break loop
			}
			bufferTimer = nil
			buffer = buffer[:0]
		}
	}
}
