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

func (f *testLogParserFixture) ReadEvent() interface{} {
	ticker := time.NewTicker(1 * time.Second)
	select {
	case event := <-f.parser.Events:
		return event
	case <-ticker.C:
		panic("no event received within 1 second")
	}
}

func (f *testLogParserFixture) Close() {
	f.parser.Close()
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
	defer f.Close()

	f.SendMessage(`Server starting with version 255`)
	assert.Equal(t, logEventServerStarting{255}, f.ReadEvent())
}
func TestLogParser_Event_LobbyConnectionFailed(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`==ERR: RegisterToLobby TCP connection failed, couldn't connect to the lobby server`)
	assert.Equal(t, logEventLobbyConnectionFailed{}, f.ReadEvent())
}

func TestLogParser_Event_LobbyConnnectionSucceeded(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`RegisterToLobby succeeded`)
	assert.Equal(t, logEventLobbyConnectionSucceeded{}, f.ReadEvent())
}

func TestLogParser_Event_NrClientsOnline(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`1 client(s) online`)
	assert.Equal(t, logEventNrClientsOnline{1}, f.ReadEvent())

	f.SendMessage(`25 client(s) online`)
	assert.Equal(t, logEventNrClientsOnline{25}, f.ReadEvent())
}

func TestLogParser_Event_Track(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`Track zandvoort_2019 was set and updated`)
	assert.Equal(t, logEventTrack{"zandvoort_2019"}, f.ReadEvent())

	f.SendMessage(`Track barcelona_2019 was set and updated`)
	assert.Equal(t, logEventTrack{"barcelona_2019"}, f.ReadEvent())
}

func TestLogParser_Event_NewConnectionRequest(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`New connection request: id 0 Max Welbezopen S76543210987654321 on car model 1`)
	assert.Equal(t, logEventNewConnectionRequest{0, "Max Welbezopen", "S76543210987654321", 1}, f.ReadEvent())

	f.SendMessage(`New connection request: id 1 Tim de Sleepwagen S12345678901234567 on car model 2`)
	assert.Equal(t, logEventNewConnectionRequest{1, "Tim de Sleepwagen", "S12345678901234567", 2}, f.ReadEvent())
}

func TestLogParser_Event_NewCarConnection(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`Creating new car connection: carId 1001, carModel 1, raceNumber #404`)
	assert.Equal(t, logEventNewCarConnection{1001, 1, 404}, f.ReadEvent())

	f.SendMessage(`Creating new car connection: carId 1002, carModel 2, raceNumber #911`)
	assert.Equal(t, logEventNewCarConnection{1002, 2, 911}, f.ReadEvent())
}

func TestLogParser_Event_DeadConnection(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`Removing dead connection 0  (last lastUdpPaketReceived 5001)`)
	assert.Equal(t, logEventDeadConnection{0}, f.ReadEvent())

	f.SendMessage(`Removing dead connection 5  (last lastUdpPaketReceived 5002)`)
	assert.Equal(t, logEventDeadConnection{5}, f.ReadEvent())
}

func TestLogParser_Event_CarRemoved(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`car 1001 has no driving connection anymore, will remove it`)
	assert.Equal(t, logEventCarRemoved{1001}, f.ReadEvent())

	f.SendMessage(`car 1013 has no driving connection anymore, will remove it`)
	assert.Equal(t, logEventCarRemoved{1013}, f.ReadEvent())
}

func TestLogParser_Event_CarPurged(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`Purging car_id 1001`)
	assert.Equal(t, logEventCarPurged{1001}, f.ReadEvent())

	f.SendMessage(`Purging car_id 1029`)
	assert.Equal(t, logEventCarPurged{1029}, f.ReadEvent())
}
