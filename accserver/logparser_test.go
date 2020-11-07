package accserver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testLogParserFixture struct {
	log    chan LogMessage
	parser *logParser
}

func newTestLogParserFixture(t *testing.T) *testLogParserFixture {
	log := make(chan LogMessage)
	return &testLogParserFixture{
		log:    log,
		parser: newLogParser(log),
	}
}

func (f *testLogParserFixture) SendMessage(msg string) {
	f.log <- LogMessage{
		Message: msg,
		Time:    time.Now(),
	}
}

func TestLogParser_CloseLogChannel(t *testing.T) {
	f := newTestLogParserFixture(t)
	close(f.log)
	_, ok := <-f.parser.Events
	assert.False(t, ok, "Events channel is not closed")
}

func TestLogParser_Close(t *testing.T) {
	f := newTestLogParserFixture(t)
	f.parser.Close()
	_, ok := <-f.parser.Events
	assert.False(t, ok, "Events channel is not closed")
}
func TestLogParser_Close_WhileAwaitingEvent(t *testing.T) {
	f := newTestLogParserFixture(t)
	f.SendMessage("RegisterToLobby succeeded")
	f.parser.Close()
	_, ok1 := <-f.parser.Events
	assert.True(t, ok1, "Missing expected message")
	_, ok2 := <-f.parser.Events
	assert.False(t, ok2, "Events channel is not closed")
}

func TestLogParser_Event_ServerStarting(t *testing.T) {
	f := newTestLogParserFixture(t)
	f.SendMessage(`Server starting with version 255`)
	assert.Equal(t, logEventServerStarting{255}, <-f.parser.Events)
}
func TestLogParser_Event_LobbyConnectionFailed(t *testing.T) {
	f := newTestLogParserFixture(t)
	f.SendMessage(`==ERR: RegisterToLobby TCP connection failed, couldn't connect to the lobby server`)
	assert.Equal(t, logEventLobbyConnectionFailed{}, <-f.parser.Events)
}

func TestLogParser_Event_LobbyConnnectionSucceeded(t *testing.T) {
	f := newTestLogParserFixture(t)
	f.SendMessage(`RegisterToLobby succeeded`)
	assert.Equal(t, logEventLobbyConnectionSucceeded{}, <-f.parser.Events)
}

func TestLogParser_Event_NrClientsOnline(t *testing.T) {
	f := newTestLogParserFixture(t)

	f.SendMessage(`1 client(s) online`)
	assert.Equal(t, logEventNrClientsOnline{1}, <-f.parser.Events)

	f.SendMessage(`25 client(s) online`)
	assert.Equal(t, logEventNrClientsOnline{25}, <-f.parser.Events)
}
