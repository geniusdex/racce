package accserver

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// LogMessage is a single message from the server log
type LogMessage struct {
	// Message is the actual message itself
	Message string
	// Time is the time at which the log message was received
	Time time.Time
}

// serverLog contains the entire server log since startup and allows reading it
type serverLog struct {
	// mutex is used to lock the channels and history during updates
	mutex *sync.Mutex
	// condMessagesAvailable is broadcasted whenever new messages are available
	condMessagesAvailable *sync.Cond
	// scanner scans the log output for new lines
	scanner *bufio.Scanner
	// history contains all past log messages
	history []LogMessage
	// isDone indicates if the process has quit
	isDone bool
	// doneChannel is closed whenever the stdout pipe of the process is closed
	doneChannel chan bool
}

const (
	initialHistoryCapacity = 1024
)

// newServerLog constructs a new serverLog object for a not-yet running process
func newServerLog(cmd *exec.Cmd) (*serverLog, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	mutex := &sync.Mutex{}

	sl := &serverLog{
		mutex:                 mutex,
		condMessagesAvailable: sync.NewCond(mutex),
		scanner:               bufio.NewScanner(stdout),
		history:               make([]LogMessage, 0, initialHistoryCapacity),
		isDone:                false,
		doneChannel:           make(chan bool),
	}

	go sl.monitor()

	return sl, nil
}

// NewChannel creates a new channel over which all server log messages will be sent
func (sl *serverLog) NewChannel() <-chan LogMessage {
	channel := make(chan LogMessage)

	go sl.feedChannel(channel)

	return channel
}

// Wait waits for the server log to be closed
func (sl *serverLog) Wait() {
	<-sl.doneChannel
}

// monitor watches the server log for new messages and handles them
func (sl *serverLog) monitor() {
	for sl.scanner.Scan() {
		sl.handleLine(strings.TrimSpace(sl.scanner.Text()))
	}
	if err := sl.scanner.Err(); err != nil {
		log.Printf("Error while reading server console: %v", err)
	}
	sl.isDone = true
	sl.condMessagesAvailable.Broadcast()
	close(sl.doneChannel)
	// "the source-monitor error, participants might misattribute"
}

// handleLine handles a new log line coming in
func (sl *serverLog) handleLine(line string) {
	msg := LogMessage{line, time.Now()}

	sl.mutex.Lock()

	// Store msg in history
	sl.history = append(sl.history, msg)

	sl.mutex.Unlock()

	sl.condMessagesAvailable.Broadcast()
}

// mutexClient wraps a mutex to allow multiple unlocking by the same client
//
// This is used by feedChannel because it defers unlocking the mutex to avoid keeping it locked in
// case of errors, but it might panic while sending messages over the channel. During sending the
// mutex is not locked by this function. Another client might lock it, and to avoid unlocking it
// in that case we need an extra guard to know if it was really this goroutine that locked it.
type mutexClient struct {
	mutex  sync.Locker
	locked bool
}

func newMutexClient(mutex sync.Locker) *mutexClient {
	return &mutexClient{
		mutex:  mutex,
		locked: false,
	}
}

func (mc *mutexClient) Lock() {
	mc.mutex.Lock()
	mc.locked = true
}

func (mc *mutexClient) Unlock() {
	if mc.locked {
		mc.locked = false
		mc.mutex.Unlock()
	}
}

// feedChannel feeds log messages to a channel
func (sl *serverLog) feedChannel(channel chan<- LogMessage) {
	messagesSent := 0

	mutex := newMutexClient(sl.mutex)
	mutex.Lock()
	defer mutex.Unlock()

	for {
		for len(sl.history) <= messagesSent {
			if sl.isDone {
				close(channel)
				return
			}
			sl.condMessagesAvailable.Wait()
		}

		// Store messages to send, then unlock the mutex before actually sending them. This way we
		// don't block the mutex when a client is misbehaving.
		messagesToSend := sl.history[messagesSent:]

		mutex.Unlock()

		for _, msg := range messagesToSend {
			channel <- msg
		}

		messagesSent += len(messagesToSend)

		mutex.Lock()
	}
}
