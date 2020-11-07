package accserver

import (
	"log"

	"github.com/geniusdex/racce/accdata"
)

// ServerState represents the current state of the server instance
type ServerState string

const (
	ServerStateOffline       ServerState = "offline"
	ServerStateStarting      ServerState = "starting"
	ServerStateNotRegistered ServerState = "not_registered"
	ServerStateOnline        ServerState = "online"
)

// LiveStateEvents contains channels for all types of events sent
//
// All channels must always be fully read until they are closed, to avoid hanging the
// event generating goroutine. All channels will be closed at the same time, but there
// is no guarantee that there are no pending messages on any other channel when one is
// closed. There is helper method Flush() to empty all channels on shutdown.
type LiveStateEvents struct {
	ServerState chan ServerState
	NrClients   chan int
	Track       chan *accdata.Track
}

// Flush reads all remaining events on all channels until they are closed.
//
// Never call this method before at least one of the event channels has been closed!
func (events *LiveStateEvents) Flush() {
	for range events.ServerState {
	}
}

// LiveState is the live state of the accServer
type LiveState struct {
	// ServerState is the current state of the server
	ServerState ServerState

	// NrClients is the number of clients currently connected to the server
	NrClients int

	// Track is the current track on the server; will never be nil
	Track *accdata.Track

	// eventListeners contains all active event listeners
	eventListeners []*LiveStateEvents

	// stopMonitoring is a channel used to indicate when the monitoring should stop
	stopMonitoring chan bool
}

func newLiveState() *LiveState {
	return &LiveState{
		ServerState: ServerStateOffline,
		NrClients:   0,
		Track:       accdata.Tracks[0],
	}
}

// NewEventChannels creates new channels for the state events
func (ls *LiveState) NewEventChannels() *LiveStateEvents {
	events := &LiveStateEvents{
		ServerState: make(chan ServerState),
		NrClients:   make(chan int),
		Track:       make(chan *accdata.Track),
	}

	ls.eventListeners = append(ls.eventListeners, events)

	return events
}

//--- Derived information ---//

// IsRunning indicates if the server is actually running (Online or NotRegistered)
func (ls *LiveState) IsRunning() bool {
	return ls.ServerState == ServerStateOnline || ls.ServerState == ServerStateNotRegistered
}

//--- State updates ---//

func (ls *LiveState) setServerState(value ServerState) {
	ls.ServerState = value
	for _, listener := range ls.eventListeners {
		listener.ServerState <- value
	}
}

func (ls *LiveState) setNrClients(value int) {
	ls.NrClients = value
	for _, listener := range ls.eventListeners {
		listener.NrClients <- value
	}
}

func (ls *LiveState) setTrack(track *accdata.Track) {
	ls.Track = track
	for _, listener := range ls.eventListeners {
		listener.Track <- track
	}
}

//--- Event reading and handling ---//

func (ls *LiveState) newInstance(logEvents <-chan interface{}) {
	if ls.stopMonitoring != nil {
		ls.stopMonitoring <- true
	}
	ls.stopMonitoring = make(chan bool)

	go ls.monitorEvents(logEvents, ls.stopMonitoring)
}

func (ls *LiveState) monitorEvents(logEvents <-chan interface{}, stopMonitoring chan bool) {
	// stopMonitoring is passed in the arguments since the one in LiveState will change when a new instance
	// is started, and we can set it to nil to indicate that we are no longer the active instance

	ls.setServerState(ServerStateStarting)
	ls.setNrClients(0)

	for logEvents != nil || stopMonitoring != nil {
		select {
		case event, ok := <-logEvents:
			if !ok {
				if stopMonitoring != nil {
					ls.setServerState(ServerStateOffline)
				}
				logEvents = nil
			} else if stopMonitoring != nil {
				ls.handleLogEvent(event)
			}

		case <-stopMonitoring:
			stopMonitoring = nil
		}
	}
}

func (ls *LiveState) handleLogEvent(event interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Unable to handle log event (%v): %v", event, r)
		}
	}()

	if e, ok := event.(logEventServerStarting); ok {
		ls.handleServerStarting(e)
	} else if e, ok := event.(logEventLobbyConnectionSucceeded); ok {
		ls.handleLobbyConnectionSucceeded(e)
	} else if e, ok := event.(logEventLobbyConnectionFailed); ok {
		ls.handleLobbyConnectionFailed(e)
	} else if e, ok := event.(logEventNrClientsOnline); ok {
		ls.handleNrClientsOnline(e)
	} else if e, ok := event.(logEventTrack); ok {
		ls.handleTrack(e)
	}
}

func (ls *LiveState) handleServerStarting(event logEventServerStarting) {
	ls.setServerState(ServerStateNotRegistered)
}

func (ls *LiveState) handleLobbyConnectionSucceeded(event logEventLobbyConnectionSucceeded) {
	ls.setServerState(ServerStateOnline)
}

func (ls *LiveState) handleLobbyConnectionFailed(event logEventLobbyConnectionFailed) {
	ls.setServerState(ServerStateNotRegistered)
}

func (ls *LiveState) handleNrClientsOnline(event logEventNrClientsOnline) {
	ls.setNrClients(event.NrClients)
}

func (ls *LiveState) handleTrack(event logEventTrack) {
	track := accdata.TrackByLabel(event.Track)
	if track != nil {
		ls.setTrack(track)
	}
}
