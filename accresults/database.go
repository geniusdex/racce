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
        "Q": "Qualifying",
        "R": "Race",
    }
)

type Event struct {
    EventId string
    TrackName string
    EndTime time.Time
    Sessions []*Session
}

type Database struct {
    Mutex *sync.RWMutex
    
    Sessions map[string]*Session
    SessionNamesSortedOnEndTime []string
    
    Players map[string]*Player
    PlayerIdsSortedOnLastName []string
    PlayerIdsSortedOnNrSessions []string
    
    Events map[string]*Event
    EventIdsSortedOnEndTime []string
    lastEvent *Event
}

func (db *Database) addSessionName(sessionName string) {
    db.SessionNamesSortedOnEndTime = append(db.SessionNamesSortedOnEndTime, sessionName)
    
    sort.Slice(db.SessionNamesSortedOnEndTime, func(i, j int) bool {
        a := db.Sessions[db.SessionNamesSortedOnEndTime[i]]
        b := db.Sessions[db.SessionNamesSortedOnEndTime[j]]
        return a.EndTime.After(b.EndTime)
    })
}

func (db* Database) getOrCreatePlayer(playerId string) *Player {
    player, ok := db.Players[playerId]
    if !ok {
        player = NewPlayer(playerId)
        db.Players[playerId] = player
        db.PlayerIdsSortedOnLastName = append(db.PlayerIdsSortedOnLastName, playerId)
        db.PlayerIdsSortedOnNrSessions = append(db.PlayerIdsSortedOnNrSessions, playerId)
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
    
    sort.Slice(db.PlayerIdsSortedOnLastName, func(i, j int) bool {
        a := db.Players[db.PlayerIdsSortedOnLastName[i]].MostRecentName
        b := db.Players[db.PlayerIdsSortedOnLastName[j]].MostRecentName
        if strings.ToLower(a.LastName) == strings.ToLower(b.LastName) {
            return strings.ToLower(a.FirstName) < strings.ToLower(b.FirstName)
        }
        return strings.ToLower(a.LastName) < strings.ToLower(b.LastName)
    })
    
    sort.Slice(db.PlayerIdsSortedOnNrSessions, func(i, j int) bool {
        a := db.Players[db.PlayerIdsSortedOnNrSessions[i]]
        b := db.Players[db.PlayerIdsSortedOnNrSessions[j]]
        return len(a.SessionNames) > len(b.SessionNames)
    })
}

func (db *Database) resolveEventForSession(session *Session) *Event {
    if db.lastEvent.TrackName != session.TrackName || session.SessionIndex == 0 {
        eventId := strings.TrimRight(session.SessionName, "_FPQR")
        db.lastEvent = &Event{eventId, session.TrackName, session.EndTime, nil}
        db.Events[eventId] = db.lastEvent
        db.EventIdsSortedOnEndTime = append(db.EventIdsSortedOnEndTime, eventId)
        
        sort.Slice(db.EventIdsSortedOnEndTime, func(i, j int) bool {
            a := db.Events[db.EventIdsSortedOnEndTime[i]]
            b := db.Events[db.EventIdsSortedOnEndTime[j]]
            return a.EndTime.After(b.EndTime)
        })
    }
    db.lastEvent.EndTime = session.EndTime
    db.lastEvent.Sessions = append(db.lastEvent.Sessions, session)
    return db.lastEvent
}


func (db *Database) addSession(sessionName string, session *Session) {
    session.SessionName = sessionName
    session.SessionTypeString = sessionTypeNames[session.SessionType]
    db.Sessions[sessionName] = session
    db.addSessionName(sessionName)
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
    session, err := LoadSessionFromFile(resultsPath + fileName, sessionTime)
    if err != nil {
        log.Print(err)
        return
    }
    
    db.Mutex.Lock()
    defer db.Mutex.Unlock()
    db.addSession(sessionName, session)
}

func (db *Database) monitorResultsDir(resultsPath string) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    err = watcher.Add(resultsPath)
    if err != nil {
        log.Fatal(err)
    }
    
    for {
        select {
        case event := <-watcher.Events:
            fileName := filepath.Base(event.Name)
            if event.Op == fsnotify.Create {
                time.AfterFunc(5 * time.Second, func() {
                    log.Printf("Loading new session file '%s'", fileName)
                    db.loadSessionFile(resultsPath, fileName)
                })
            }
        case err := <-watcher.Errors:
            log.Fatal(err)
        }
    }
}

func LoadDatabase(resultsPath string) (*Database, error) {
    var db = &Database{
        &sync.RWMutex{},
        make(map[string]*Session),
        nil,
        make(map[string]*Player),
        nil,
        nil,
        make(map[string]*Event),
        nil,
        &Event{"__", "__", time.Now(), nil},
    }
    
    resultsPath = strings.TrimRight(resultsPath, "/") + "/"
    
    files, err := ioutil.ReadDir(resultsPath)
    if err != nil {
        return nil, err
    }
    
    sort.Slice(files, func(i, j int) bool {
        return files[i].Name() < files[j].Name()
    })
    
    for _, f := range files {
        fileName := f.Name()
        db.loadSessionFile(resultsPath, fileName)
    }
    
    go db.monitorResultsDir(resultsPath)
    
    return db, err
}
