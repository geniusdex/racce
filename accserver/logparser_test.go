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

func TestLogParser_Event_SessionPhaseChanged(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`Detected sessionPhase <session> -> <session> (Qualifying)`)
	assert.Equal(t, logEventSessionPhaseChanged{"Qualifying", "session"}, f.ReadEvent())

	f.SendMessage(`Detected sessionPhase <waiting for drivers> -> <pre session> (Race)`)
	assert.Equal(t, logEventSessionPhaseChanged{"Race", "pre session"}, f.ReadEvent())
}

func TestLogParser_Event_ResettingWeekend(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`Resetting race weekend`)
	assert.Equal(t, logEventResettingWeekend{}, f.ReadEvent())
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

func TestLogParser_Event_NewLapTime(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	// These matches were used for laptimes previously, and should no longer match to avoid counting laps twice
	f.SendMessage(`New laptime: 142474 for carId 1002 with lapstates: HasCut, IsOutLap, IsInLap,  (raw 13)`)
	f.SendMessage(`New laptime: 107466 for carId 1004 with lapstates: HasCut,  (raw 1)`)
	f.SendMessage(`New laptime: 107142 for carId 1005 with lapstates:  (raw 0)`)

	// These lines should not trigger an update
	f.SendMessage(`Lap  carId 1020, driverId 0, lapTime 35791:23:647, timestampMS 2192453.000000, flags: %d0, S1 4:11:730, S2 0:40:917, S3 0:34:929, fuel 0.000000`)
	f.SendMessage(`Lap carId 1003, driverId 0, lapTime 35791:23:647, timestampMS 3558634.000000, flags: %d0, S1 0:30:777, fuel 0.000000`)
	f.SendMessage(`Lap  carId 1012, driverId 0, lapTime 35791:23:647, timestampMS 3571140.000000, flags: %d1, S1 0:30:360, S2 0:41:811, fuel 0.000000, hasCut `)
	f.SendMessage(`Lap  carId 1009, driverId 0, lapTime 35791:23:647, timestampMS 3885634.000000, flags: %d4, S1 10:00:510, S2 0:42:834, S3 0:36:777, fuel 0.000000, OutLap `)
	f.SendMessage(`Lap  carId 1003, driverId 0, lapTime 35791:23:647, timestampMS 1202339.000000, flags: 00, S1 0:40:380, S2 1:01:878, S3 0:28:749, fuel 0.000000`)

	// These lines do
	f.SendMessage(`Lap carId 1020, driverId 0, lapTime 5:27:576, timestampMS 2192453.000000, flags: %d0, S1 4:11:730, S2 0:40:917, S3 0:34:929, fuel 56.000000`)
	assert.Equal(t, logEventNewLapTime{1020, 327576, 2192453, 0}, f.ReadEvent())

	f.SendMessage(`Lap carId 1009, driverId 0, lapTime 11:20:121, timestampMS 3885634.000000, flags: %d4, S1 10:00:510, S2 0:42:834, S3 0:36:777, fuel 22.000000, OutLap `)
	assert.Equal(t, logEventNewLapTime{1009, 680121, 3885634, 4}, f.ReadEvent())

	f.SendMessage(`Lap carId 1046, driverId 0, lapTime 1:46:830, timestampMS 4004213.000000, flags: %d1025, S1 0:29:832, S2 0:40:917, S3 0:36:081, fuel 73.000000, hasCut , SessionOver`)
	assert.Equal(t, logEventNewLapTime{1046, 106830, 4004213, 1025}, f.ReadEvent())

	f.SendMessage(`Lap carId 1003, driverId 0, lapTime 2:11:007, timestampMS 1202340.000000, flags: 00, 1 0:40:380, S2 1:01:878, S3 0:28:749, fuel 23.000000`)
	assert.Equal(t, logEventNewLapTime{1003, 131007, 1202340, 0}, f.ReadEvent())

	f.SendMessage(`Lap carId 1020, driverId 0, lapTime 2:35:097, timestampMS 2033151.000000, flags: 01, S1 0:39:060, S2 1:11:364, S3 0:44:673, fuel 3.000000, hasCut`)
	assert.Equal(t, logEventNewLapTime{1020, 155097, 2033151, 1}, f.ReadEvent())
}

func TestLogParser_Event_GridPosition(t *testing.T) {
	f := newTestLogParserFixture(t)
	defer f.Close()

	f.SendMessage(`   Car 1024 Pos 1`)
	assert.Equal(t, logEventGridPosition{1024, 1}, f.ReadEvent())

	f.SendMessage(`   Car 1016 Pos 21`)
	assert.Equal(t, logEventGridPosition{1016, 21}, f.ReadEvent())
}
