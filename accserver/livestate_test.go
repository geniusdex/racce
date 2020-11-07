package accserver

import (
	"log"
	"os"
	"testing"

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
	<-f.events.ServerState
	<-f.events.NrClients

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
