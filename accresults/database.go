package accresults

import (
    "io/ioutil"
    "log"
    "sort"
    "strings"
    "time"
)

var (
    sessionTypeNames = map[SessionType]string{
        "FP": "Free Practice",
        "Q": "Qualifying",
        "R": "Race",
    }
)

type Event struct {
    TrackName string
    EndTime time.Time
    Sessions []*Session
}

type Database struct {
    Sessions map[string]*Session
    SessionNamesSortedOnEndTime []string
    
    Players map[string]*Player
    PlayerIdsSortedOnLastName []string
    
    Events []*Event
}

func (db* Database) getOrCreatePlayer(playerId string) *Player {
    player, ok := db.Players[playerId]
    if !ok {
        player = &Player{}
        player.PlayerId = playerId
        db.Players[playerId] = player
        db.PlayerIdsSortedOnLastName = append(db.PlayerIdsSortedOnLastName, playerId)
    }
    return player
}

func (db *Database) postprocess() {
    for sessionName, session := range db.Sessions {
        db.SessionNamesSortedOnEndTime = append(db.SessionNamesSortedOnEndTime, sessionName)
        session.SessionName = sessionName
        session.SessionTypeString = sessionTypeNames[session.SessionType]
        for _, line := range session.SessionResult.LeaderBoardLines {
            for _, driver := range line.Car.Drivers {
                player := db.getOrCreatePlayer(driver.PlayerId)
                player.mergeDriver(driver)
                player.SessionNames = append(player.SessionNames, sessionName)
            }
        }
    }
    
    sort.Slice(db.SessionNamesSortedOnEndTime, func(i, j int) bool {
        a := db.Sessions[db.SessionNamesSortedOnEndTime[i]]
        b := db.Sessions[db.SessionNamesSortedOnEndTime[j]]
        return a.EndTime.After(b.EndTime)
    })
    
    sort.Slice(db.PlayerIdsSortedOnLastName, func(i, j int) bool {
        a := db.Players[db.PlayerIdsSortedOnLastName[i]].MostRecentName
        b := db.Players[db.PlayerIdsSortedOnLastName[j]].MostRecentName
        if strings.ToLower(a.LastName) == strings.ToLower(b.LastName) {
            return strings.ToLower(a.FirstName) < strings.ToLower(b.FirstName)
        }
        return strings.ToLower(a.LastName) < strings.ToLower(b.LastName)
    })
    
    event := &Event{"__", time.Now(), nil}
    for i := len(db.SessionNamesSortedOnEndTime)-1; i >= 0; i-- {
        session := db.Sessions[db.SessionNamesSortedOnEndTime[i]]
        if event.TrackName != session.TrackName || session.SessionIndex == 0 {
            event = &Event{session.TrackName, session.EndTime, nil}
            db.Events = append(db.Events, event)
        }
        event.EndTime = session.EndTime
        event.Sessions = append(event.Sessions, session)
    }
    for i, j := 0, len(db.Events)-1; i < j; i, j = i+1, j-1 {
        db.Events[i], db.Events[j] = db.Events[j], db.Events[i]
    }
}

func parseTimeFromSessionName(name string) time.Time {
    result, err := time.ParseInLocation("060102_150405", strings.TrimRight(name, "_FPQR"), time.Local)
    if err != nil {
        log.Printf("Cannot parse time from session name '%s': %v", name, err)
        return time.Unix(0, 0)
    }
    return result
}

func LoadDatabase(resultsPath string) (*Database, error) {
    var db = &Database{
        make(map[string]*Session),
        nil,
        make(map[string]*Player),
        nil,
        make([]*Event, 0),
    }
    
    files, err := ioutil.ReadDir(resultsPath)
    if err != nil {
        return nil, err
    }
    
    for _, f := range files {
        fileName := f.Name()
        sessionName := strings.TrimSuffix(fileName, ".json")
        sessionTime := parseTimeFromSessionName(sessionName)
        db.Sessions[sessionName], err = LoadSessionFromFile(resultsPath + fileName, sessionTime)
        if err != nil {
            return db, err
        }
    }
    
    db.postprocess()
    
    return db, err
}
