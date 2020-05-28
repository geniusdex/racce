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

// Configuration contains the configuration options for the database
type Configuration struct {
	// ResultsDir is the directory which contains the results files
	ResultsDir string `json:"resultsDir"`
	// NewFileDelay is the number of seconds to wait after a new file appears before reading it
	NewFileDelay int `json:"newFileDelay"`
}

// resultsDir returns the ResultsDir with a single slash at the end
func (c *Configuration) resultsDir() string {
	return strings.TrimRight(c.ResultsDir, "/") + "/"
}

type Event struct {
	EventId   string
	TrackName string
	EndTime   time.Time
	Sessions  []*Session
}

type Database struct {
	Mutex *sync.RWMutex

	Sessions map[string]*Session

	Players map[string]*Player

	Events    map[string]*Event
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
		log.Print(err)
		return
	}

	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	db.addSession(sessionName, session)
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
					log.Printf("Loading new session file '%s'", fileName)
					db.loadSessionFile(config.resultsDir(), fileName)
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
		db.loadSessionFile(config.resultsDir(), fileName)
	}

	go db.monitorResultsDir(config)

	return db, err
}
