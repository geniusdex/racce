package accserver

import (
	"log"
	"regexp"
	"strconv"
)

// logEventServerStarting is the first event on server startup
type logEventServerStarting struct {
	Version int
}

// logEventLobbyConnectionFailed indicates the connection to the lobby server failed
type logEventLobbyConnectionFailed struct {
}

// logEventLobbyConnectionSucceeded indicates the connection to the lobby server succeeded
type logEventLobbyConnectionSucceeded struct {
}

// logEventNrClientsOnline contains updates to the number of online clients (drivers + spectators)
type logEventNrClientsOnline struct {
	NrClients int
}

// logEventTrack indicates which track is currently being used
type logEventTrack struct {
	Track string
}

// logEventSessionPhaseChanged indicates a change to session type or phasse
type logEventSessionPhaseChanged struct {
	Type  string
	Phase string
}

// logEventNewConnectionRequest is sent whenever a new driver connection is made
type logEventNewConnectionRequest struct {
	ConnectionID int
	PlayerName   string
	SteamID      string
	CarModelID   int
}

// logEventNewCarConnection indicates the car for a new connection
type logEventNewCarConnection struct {
	CarID      int
	CarModelID int
	RaceNumber int
}

// logEventDeadConnection is sent when a connection with a client is dead
type logEventDeadConnection struct {
	ConnectionID int
}

// logEventCarRemoved is sent when there are no more drivers connected for a car (it is no longer physically present)
type logEventCarRemoved struct {
	CarID int
}

// logEventCarPurged is sent when a car ID is being retired, i.e. when resetting the race weekend and no driver is connected for the car
type logEventCarPurged struct {
	CarID int
}

// logEventNewLapTime indicates a lap was completed by a car
type logEventNewLapTime struct {
	CarID     int
	LapTimeMS int
	// Flags as binary bitfield with 1=HasCut, 4=IsOutLap, 8=IsInLap (flagLap* constants)
	Flags int
}

// logEventGridPosition is sent at the end of qualifying when the grid positions for the race are known
type logEventGridPosition struct {
	CarID    int
	Position int
}

const (
	flagLapHasCut   = 1
	flagLapIsOutLap = 4
	flagLapIsInLap  = 8
)

type logMatcher struct {
	matcher *regexp.Regexp
	handler func([]string) interface{}
}

func newLogMatcher(expr string, handler func([]string) interface{}) *logMatcher {
	return &logMatcher{
		matcher: regexp.MustCompile(expr),
		handler: handler,
	}
}

func intOrPanic(str string) int {
	value, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return value
}

func makeLogMatchers() (ret []*logMatcher) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Failed to create log matchers: %v", r)
			ret = []*logMatcher{}
		}
	}()

	return []*logMatcher{
		newLogMatcher(
			`^Server starting with version ([0-9]+)$`,
			func(matches []string) interface{} { return logEventServerStarting{intOrPanic(matches[1])} }),
		newLogMatcher(
			`RegisterToLobby TCP connection failed`,
			func([]string) interface{} { return logEventLobbyConnectionFailed{} }),
		newLogMatcher(
			`RegisterToLobby succeeded`,
			func([]string) interface{} { return logEventLobbyConnectionSucceeded{} }),
		newLogMatcher(
			`^([0-9]+) client\(s\) online$`,
			func(matches []string) interface{} { return logEventNrClientsOnline{intOrPanic(matches[1])} }),
		newLogMatcher(
			`^Track ([a-zA-Z0-9_]+) was set and updated$`,
			func(matches []string) interface{} { return logEventTrack{matches[1]} }),
		newLogMatcher(
			`^Detected sessionPhase <([A-Za-z ]+)> -> <([A-Za-z ]+)> \(([A-Za-z ]+)\)$`,
			func(matches []string) interface{} { return logEventSessionPhaseChanged{matches[3], matches[2]} }),
		newLogMatcher(
			`^New connection request: id (\d+) (.+) (S\d+) on car model (\d+)$`,
			func(matches []string) interface{} {
				return logEventNewConnectionRequest{intOrPanic(matches[1]), matches[2], matches[3], intOrPanic(matches[4])}
			}),
		newLogMatcher(
			`^Creating new car connection: carId (\d+), carModel (\d+), raceNumber #(\d+)$`,
			func(matches []string) interface{} {
				return logEventNewCarConnection{intOrPanic(matches[1]), intOrPanic(matches[2]), intOrPanic(matches[3])}
			}),
		newLogMatcher(
			`Removing dead connection (\d+)`,
			func(matches []string) interface{} { return logEventDeadConnection{intOrPanic(matches[1])} }),
		newLogMatcher(
			`^car (\d+) has no driving connection anymore, will remove it$`,
			func(matches []string) interface{} { return logEventCarRemoved{intOrPanic(matches[1])} }),
		newLogMatcher(
			`^Purging car_id (\d+)$`,
			func(matches []string) interface{} { return logEventCarPurged{intOrPanic(matches[1])} }),
		newLogMatcher(
			`^New laptime: (\d+) for carId (\d+) with lapstates: [a-zA-Z, ]* \(raw (\d+)\)$`,
			func(matches []string) interface{} {
				return logEventNewLapTime{intOrPanic(matches[2]), intOrPanic(matches[1]), intOrPanic(matches[3])}
			}),
		newLogMatcher(
			`^\s*Car (\d+) Pos (\d+)$`,
			func(matches []string) interface{} {
				return logEventGridPosition{intOrPanic(matches[1]), intOrPanic(matches[2])}
			}),
	}
}

type logParser struct {
	Events   chan interface{}
	close    chan bool
	matchers []*logMatcher
}

func newLogParser(logChannel <-chan LogMessage) *logParser {
	parser := &logParser{
		Events:   make(chan interface{}),
		close:    make(chan bool, 1),
		matchers: makeLogMatchers(),
	}

	go parser.parseLogMessages(logChannel)

	return parser
}

func (parser *logParser) Close() {
	parser.close <- true
}

func (parser *logParser) parseMessage(msg string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error (%v) while parsing log line `%v`", msg, r)
		}
	}()

	for _, matcher := range parser.matchers {
		matches := matcher.matcher.FindStringSubmatch(msg)
		if matches != nil {
			parser.Events <- matcher.handler(matches)
			return
		}
	}
}

func (parser *logParser) parseLogMessages(logChannel <-chan LogMessage) {
loop:
	for {
		select {
		case msg, ok := <-logChannel:
			if !ok {
				break loop
			}

			parser.parseMessage(msg.Message)

		case <-parser.close:
			break loop
		}
	}

	close(parser.Events)
}
