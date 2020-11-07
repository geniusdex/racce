package accserver

import (
	"log"
	"regexp"
	"strconv"
)

type logEventServerStarting struct {
	Version int
}

type logEventLobbyConnectionFailed struct {
}

type logEventLobbyConnectionSucceeded struct {
}

type logEventNrClientsOnline struct {
	NrClients int
}

type logEventTrack struct {
	Track string
}

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
