package accresults

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	sessionTypeNames = map[SessionType]string{
		"FP": "Free Practice",
		"Q":  "Qualifying",
		"R":  "Race",
	}
)

// Options contains the options which influence result parsing and interpretation
type Options struct {
	// FilterCarsWithoutLaps enables filtering out cars without any completed lap
	FilterCarsWithoutLaps bool `json:"filterCarsWithoutLaps"`
	// filterSessionsWithoutCars enables filteing out sessions without any cars
	FilterSessionsWithoutCars bool `json:"filterSessionsWithoutCars"`
}

// Configuration contains the configuration options for the database
type Configuration struct {
	// ResultsDir is the directory which contains the results files
	ResultsDir string
	// NewFileDelay is the number of seconds to wait after a new file appears before reading it
	NewFileDelay int
	// Options contains the options which influence result parsing and interpretation
	Options Options
}

// resultsDir returns the ResultsDir with a single slash at the end
func (c *Configuration) resultsDir() string {
	return strings.TrimRight(c.ResultsDir, "/") + "/"
}

// Event identifies an event consisting of one or more sessions
type Event struct {
	EventId   string
	TrackName string
	EndTime   time.Time
	Sessions  []*Session
}

// Database contains the results and derived data as obtained from parsing the result files
type Database struct {
	// options contains the options which influence result parsing and interpretation
	options Options

	// Mutex must be locked whenever data is read from the database, to avoid database updates interfering with the read
	Mutex *sync.RWMutex

	// Sessions contains all individual sessions keyed on the session ID
	Sessions map[string]*Session

	// Players contains all the players which took part in a raw keyed on player ID
	Players map[string]*Player

	// Events contains all events keyed on event ID
	Events map[string]*Event
	// lastEvent is the last event that was added to the database
	lastEvent *Event
}

func (db *Database) getOrCreatePlayer(playerId string) *Player {
	player, ok := db.Players[playerId]
	if !ok {
		player = NewPlayer(playerId)
		db.Players[playerId] = player
	}
	return player
}

func (db *Database) resolvePlayersInSession(sessionName string, session *Session, event *Event) {
	for _, line := range session.SessionResult.LeaderBoardLines {
		for _, driver := range line.Car.Drivers {
			player := db.getOrCreatePlayer(driver.PlayerId)
			player.mergeDriver(driver)
			player.SessionNames = append(player.SessionNames, sessionName)
			player.Events[event.EventId] = event
		}
	}

}

func (db *Database) resolveEventForSession(session *Session) *Event {
	if db.lastEvent.TrackName != session.TrackName || session.SessionIndex == 0 {
		eventId := strings.TrimRight(session.SessionName, "_FPQR")
		db.lastEvent = &Event{eventId, session.TrackName, session.EndTime, nil}
		db.Events[eventId] = db.lastEvent
	}
	db.lastEvent.EndTime = session.EndTime
	db.lastEvent.Sessions = append(db.lastEvent.Sessions, session)
	return db.lastEvent
}

func (db *Database) addSession(sessionName string, session *Session) {
	session.SessionName = sessionName
	session.SessionTypeString = sessionTypeNames[session.SessionType]
	db.Sessions[sessionName] = session
	event := db.resolveEventForSession(session)
	db.resolvePlayersInSession(sessionName, session, event)
}

func isSessionFile(fileName string) bool {
	return strings.HasSuffix(fileName, "_FP.json") ||
		strings.HasSuffix(fileName, "_Q.json") ||
		strings.HasSuffix(fileName, "_R.json")
}

func parseTimeFromSessionName(name string) time.Time {
	result, err := time.ParseInLocation("060102_150405", strings.TrimRight(name, "_FPQR"), time.Local)
	if err != nil {
		log.Printf("Cannot parse time from session name '%s': %v", name, err)
		return time.Unix(0, 0)
	}
	return result
}

func (db *Database) loadSessionFile(resultsPath string, fileName string) {
	sessionName := strings.TrimSuffix(fileName, ".json")
	sessionTime := parseTimeFromSessionName(sessionName)
	session, err := LoadSessionFromFile(resultsPath+fileName, sessionTime)
	if err != nil {
		log.Printf("Error loading session results file '%v': %v", fileName, err)
		return
	}

	db.applyFiltersToSession(session)
	if !db.isSessionFiltered(session) {
		db.Mutex.Lock()
		defer db.Mutex.Unlock()
		db.addSession(sessionName, session)
	}
}

func (db *Database) applyFiltersToSession(session *Session) {
	if db.options.FilterCarsWithoutLaps {
		session.filterCarsWithoutLaps()
	}
}

func (db *Database) isSessionFiltered(session *Session) bool {
	return len(session.SessionResult.LeaderBoardLines) == 0
}

func (db *Database) monitorResultsDir(config *Configuration) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add(config.resultsDir())
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Aborting watcher due to error: %v", r)
		}
		log.Print("Closing results dir watcher...")
		if err := watcher.Close(); err != nil {
			log.Printf("Error closing results dir watcher: %v", err)
		}
	}()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			fileName := filepath.Base(event.Name)
			if event.Op&fsnotify.Create == fsnotify.Create {
				time.AfterFunc(time.Duration(config.NewFileDelay)*time.Second, func() {
					if isSessionFile(fileName) {
						log.Printf("Loading new session file '%s'", fileName)
						db.loadSessionFile(config.resultsDir(), fileName)
					} else {
						log.Printf("Ignoring file '%s' because it is not a session results file", fileName)
					}
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Panic(err)
		}
	}
}

// LoadDatabase loads a database from disk and starts monitoring it
func LoadDatabase(config *Configuration) (*Database, error) {
	var db = &Database{
		config.Options,
		&sync.RWMutex{},
		make(map[string]*Session),
		make(map[string]*Player),
		make(map[string]*Event),
		&Event{"__", "__", time.Now(), nil},
	}

	files, err := ioutil.ReadDir(config.resultsDir())
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, f := range files {
		fileName := f.Name()
		if isSessionFile(fileName) {
			db.loadSessionFile(config.resultsDir(), fileName)
		} else {
			log.Printf("Ignoring file '%s' because it is not a session results file", fileName)
		}
	}

	go db.monitorResultsDir(config)

	return db, err
}
