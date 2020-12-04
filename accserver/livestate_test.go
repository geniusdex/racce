package accserver

import (
	"log"
	"os"
	"testing"

	"github.com/geniusdex/racce/accdata"

	"net/http"
	_ "net/http/pprof"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	os.Exit(m.Run())
}

type testLiveStateFixture struct {
	assert    *assert.Assertions
	state     *LiveState
	logEvents chan interface{}
	events    *LiveStateEvents
}

func newTestLiveStateFixture(t *testing.T) *testLiveStateFixture {
	state := newLiveState()

	f := &testLiveStateFixture{
		assert:    assert.New(t),
		state:     state,
		logEvents: make(chan interface{}),
		events:    state.NewEventChannels(),
	}

	f.state.newInstance(f.logEvents)
	// Eat initial events always sent out
	<-f.events.ServerState
	<-f.events.NrClients

	// Do some lookups to avoid cache building during first use
	accdata.TrackByLabel("zandvoort")
	accdata.CarModelByID(58)

	return f
}

//--- ServerState ---//

func TestLiveState_ServerState_Lifecycle(t *testing.T) {
	assert := assert.New(t)

	state := newLiveState()
	assert.Equal(ServerStateOffline, state.ServerState)
	assert.False(state.IsRunning())

	events := state.NewEventChannels()

	logEvents := make(chan interface{})
	state.newInstance(logEvents)
	assert.Equal(ServerStateStarting, <-events.ServerState)
	assert.Equal(ServerStateStarting, state.ServerState)
	assert.False(state.IsRunning())
	assert.Equal(0, <-events.NrClients)
	assert.Equal(0, state.NrClients)

	logEvents <- logEventServerStarting{Version: 0}
	assert.Equal(ServerStateNotRegistered, <-events.ServerState)
	assert.Equal(ServerStateNotRegistered, state.ServerState)
	assert.True(state.IsRunning())

	logEvents <- logEventLobbyConnectionSucceeded{}
	assert.Equal(ServerStateOnline, <-events.ServerState)
	assert.Equal(ServerStateOnline, state.ServerState)
	assert.True(state.IsRunning())

	close(logEvents)
	assert.Equal(ServerStateOffline, <-events.ServerState)
	assert.Equal(ServerStateOffline, state.ServerState)
	assert.False(state.IsRunning())
}

func TestLiveState_ServerState_NewInstance(t *testing.T) {
	f := newTestLiveStateFixture(t)

	// Server state should only respond to logEvents2 from now on
	logEvents2 := make(chan interface{})
	f.state.newInstance(logEvents2)
	assert.Equal(t, ServerStateStarting, <-f.events.ServerState)
	assert.Equal(t, 0, <-f.events.NrClients)

	f.logEvents <- logEventServerStarting{Version: 0}
	assert.Equal(t, ServerStateStarting, f.state.ServerState)

	logEvents2 <- logEventLobbyConnectionSucceeded{}
	assert.Equal(t, ServerStateOnline, <-f.events.ServerState)

	close(f.logEvents)
	assert.Equal(t, ServerStateOnline, f.state.ServerState)
}

func TestLiveState_ServerState_LobbyConnectionLost(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventLobbyConnectionSucceeded{}
	assert.Equal(t, ServerStateOnline, <-f.events.ServerState)

	f.logEvents <- logEventLobbyConnectionFailed{}
	assert.Equal(t, ServerStateNotRegistered, <-f.events.ServerState)

	f.logEvents <- logEventLobbyConnectionSucceeded{}
	assert.Equal(t, ServerStateOnline, <-f.events.ServerState)
}

//--- NrClients ---//
func TestLiveState_NrClients(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventNrClientsOnline{5}
	assert.Equal(t, 5, <-f.events.NrClients)
	assert.Equal(t, 5, f.state.NrClients)

	f.logEvents <- logEventNrClientsOnline{0}
	assert.Equal(t, 0, <-f.events.NrClients)
	assert.Equal(t, 0, f.state.NrClients)
}

//--- Track ---//
func TestLiveState_Track(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventTrack{"brands_hatch"}
	assert.Equal(t, accdata.TrackByLabel("brands_hatch"), <-f.events.Track)
	assert.Equal(t, accdata.TrackByLabel("brands_hatch"), f.state.Track)
}

//--- Session State ---//
func TestLiveState_SessionPhase(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventSessionPhaseChanged{"Qualifying", "session"}
	assert.Equal(t, &SessionState{SessionTypeQualifying, SessionPhaseSession}, <-f.events.SessionState)
	assert.Equal(t, SessionTypeQualifying, f.state.SessionState.Type)
	assert.Equal(t, SessionPhaseSession, f.state.SessionState.Phase)
}

//--- Car Updates ---//
func TestLiveState_NewCar(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventNewConnectionRequest{5, "Driver One", "S76543210987654321", 1}
	f.logEvents <- logEventNewCarConnection{1001, 1, 404}
	driver := &Driver{
		ConnectionID: 5,
		Name:         "Driver One",
		PlayerID:     "S76543210987654321",
	}
	carState := &CarState{
		CarID:         1001,
		RaceNumber:    404,
		CarModel:      accdata.CarModelByID(1),
		Drivers:       []*Driver{driver},
		CurrentDriver: driver,
		Position:      1,
	}
	assert.Equal(t, carState, <-f.events.CarState)
	assert.Equal(t, carState, f.state.CarState[1001])

	carState.Drivers = []*Driver{}
	f.logEvents <- logEventDeadConnection{5}
	assert.Equal(t, carState, <-f.events.CarState)
	assert.Equal(t, carState, f.state.CarState[1001])
}

func TestLiveState_CarPurged(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventNewConnectionRequest{6, "Driver One", "S76543210987654321", 5}
	f.logEvents <- logEventNewCarConnection{1002, 5, 42}
	assert.NotNil(t, <-f.events.CarState)
	assert.NotNil(t, f.state.CarState[1002])
	assert.Equal(t, 1, f.state.CarState[1002].Position)

	f.logEvents <- logEventNewConnectionRequest{7, "Driver Two", "S5", 6}
	f.logEvents <- logEventNewCarConnection{1004, 6, 37}
	assert.NotNil(t, <-f.events.CarState)
	assert.NotNil(t, f.state.CarState[1004])
	assert.Equal(t, 2, f.state.CarState[1004].Position)

	f.logEvents <- logEventCarPurged{1002}
	assert.Equal(t, 1002, <-f.events.CarPurged)
	assert.Nil(t, f.state.CarState[1002])
	carState := <-f.events.CarState
	assert.Equal(t, 1004, carState.CarID)
	assert.Equal(t, 1, carState.Position)
	assert.Equal(t, carState, f.state.CarState[1004])
}

func TestLiveState_CarsWithoutDriversPurgedWhenSessionTypeChanges(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventSessionPhaseChanged{"Qualifying", "session"}
	<-f.events.SessionState

	f.logEvents <- logEventNewConnectionRequest{6, "Driver One", "S76543210987654321", 5}
	f.logEvents <- logEventNewCarConnection{1002, 5, 42}
	<-f.events.CarState

	f.logEvents <- logEventNewConnectionRequest{7, "Driver Two", "S5", 6}
	f.logEvents <- logEventNewCarConnection{1004, 6, 37}
	<-f.events.CarState

	f.logEvents <- logEventDeadConnection{7}
	<-f.events.CarState

	f.logEvents <- logEventSessionPhaseChanged{"Race", "session"}
	<-f.events.SessionState
	go func() { <-f.events.CarState }() // Eat car state update vor 1002
	assert.Equal(t, 1004, <-f.events.CarPurged)
	assert.Nil(t, f.state.CarState[1004])
	assert.NotNil(t, f.state.CarState[1002])
}

func TestLiveState_GridPosition(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventNewConnectionRequest{6, "Driver One", "S76543210987654321", 5}
	f.logEvents <- logEventNewCarConnection{1002, 5, 42}
	assert.NotNil(t, <-f.events.CarState)

	f.logEvents <- logEventGridPosition{1002, 6}
	carState := <-f.events.CarState
	assert.Equal(t, 6, carState.Position)
	assert.Equal(t, 6, f.state.CarState[1002].Position)
}

//--- Lap times ---//
func TestLiveState_NewLapTime(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventNewConnectionRequest{6, "Driver One", "S76543210987654321", 5}
	f.logEvents <- logEventNewCarConnection{1002, 5, 42}
	assert.NotNil(t, <-f.events.CarState)

	f.logEvents <- logEventNewLapTime{1002, 123456, 0}
	carState := <-f.events.CarState
	assert.Equal(t, 123456, carState.BestLapMS)
	assert.Equal(t, carState, f.state.CarState[1002])

	f.logEvents <- logEventNewLapTime{1002, 123457, 0}
	f.logEvents <- logEventNewLapTime{1002, 123000, 1}
	f.logEvents <- logEventNewLapTime{1002, 123000, 4}
	f.logEvents <- logEventNewLapTime{1002, 123000, 8}
	f.logEvents <- logEventNewLapTime{1002, 123000, 13}

	f.logEvents <- logEventNewLapTime{1002, 123400, 0}
	carState = <-f.events.CarState
	assert.Equal(t, 123400, carState.BestLapMS)
	assert.Equal(t, carState, f.state.CarState[1002])
}

func TestLiveState_LapTimesRemovedWhenSessionTypeChanges(t *testing.T) {
	f := newTestLiveStateFixture(t)

	f.logEvents <- logEventSessionPhaseChanged{"Qualifying", "session"}
	<-f.events.SessionState

	f.logEvents <- logEventNewConnectionRequest{6, "Driver One", "S76543210987654321", 5}
	f.logEvents <- logEventNewCarConnection{1002, 5, 42}
	assert.NotNil(t, <-f.events.CarState)

	f.logEvents <- logEventNewLapTime{1002, 123456, 0}
	<-f.events.CarState
	assert.Equal(t, 123456, f.state.CarState[1002].BestLapMS)

	f.logEvents <- logEventSessionPhaseChanged{"Race", "session"}
	<-f.events.SessionState
	carState := <-f.events.CarState
	assert.Equal(t, 0, carState.BestLapMS)
	assert.Equal(t, 0, f.state.CarState[1002].BestLapMS)
}
