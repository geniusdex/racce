package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
    "strings"
    "strconv"
    "github.com/geniusdex/racce/accresults"
)

var (
    accdb *accresults.Database
)

func templateFunctionMap() template.FuncMap{
    return template.FuncMap{
        "add": func(a, b int) int {
            return a + b
        },
        "div": func(a, b int) float64 {
            return float64(a) / float64(b)
        },
        "laptime": func(time int) string {
            milliseconds := time % 1000
            time = time / 1000
            seconds := time % 60
            minutes := time / 60
            
            return fmt.Sprintf("%d:%02d.%03d", minutes, seconds, milliseconds)
        },
        "carInSession": func(carId int, session *accresults.Session) *accresults.Car {
            return session.FindCarById(carId)
        },
        "contains": func(haystack []string, needle string) bool {
            for _, element := range haystack {
                if element == needle {
                    return true
                }
            }
            return false
        },
    }
}

func executeTemplate(w http.ResponseWriter, name string, data interface{}) {
    t,err := template.New("templates").Funcs(templateFunctionMap()).ParseGlob("templates/*.html")
    if err != nil {
        log.Print(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    
    err = t.ExecuteTemplate(w, name, data)
    if err != nil {
        log.Print(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    accdb.Mutex.RLock()
    defer accdb.Mutex.RUnlock()
    
    executeTemplate(w, "index.html", accdb)
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
    pathComponents := strings.Split(r.URL.Path, "/")
    if len(pathComponents) != 3 {
        log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    accdb.Mutex.RLock()
    defer accdb.Mutex.RUnlock()
    
    eventId := pathComponents[2]
    event, ok := accdb.Events[eventId]
    if !ok {
        log.Printf("Event '%s' not found in database", eventId)
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    executeTemplate(w, "event.html", event)
}

type playerPage struct {
    Database *accresults.Database
    Player *accresults.Player
}

func playerHandler(w http.ResponseWriter, r *http.Request) {
    pathComponents := strings.Split(r.URL.Path, "/")
    if len(pathComponents) != 3 {
        log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    accdb.Mutex.RLock()
    defer accdb.Mutex.RUnlock()
    
    playerId := pathComponents[2]
    player, ok := accdb.Players[playerId]
    if !ok {
        log.Printf("Player '%s' not found in database", playerId)
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    executeTemplate(w, "player.html", &playerPage{accdb, player})
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
    pathComponents := strings.Split(r.URL.Path, "/")
    if len(pathComponents) != 3 {
        log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    accdb.Mutex.RLock()
    defer accdb.Mutex.RUnlock()
    
    sessionName := pathComponents[2]
    session, ok := accdb.Sessions[sessionName]
    if !ok {
        log.Printf("Session '%s' not found in database", sessionName)
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    executeTemplate(w, "session.html", session)
}

type sessionCarPage struct {
    Session *accresults.Session
    Car *accresults.Car
}

func sessionCarHandler(w http.ResponseWriter, r *http.Request) {
    pathComponents := strings.Split(r.URL.Path, "/")
    if len(pathComponents) != 4 {
        log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    accdb.Mutex.RLock()
    defer accdb.Mutex.RUnlock()
    
    sessionName := pathComponents[2]
    session, ok := accdb.Sessions[sessionName]
    if !ok {
        log.Printf("Session '%s' not found in database", sessionName)
        w.WriteHeader(http.StatusNotFound)
        return
    }
    carId,err := strconv.Atoi(pathComponents[3])
    if err != nil {
        log.Printf("Car ID '%s' is not numeric", carId)
        w.WriteHeader(http.StatusNotFound)
        return
    }
    car := session.FindCarById(carId)
    if car == nil {
        log.Printf("Car ID '%d' not present in session", carId)
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    executeTemplate(w, "sessioncar.html", &sessionCarPage{session, car})
}

func RunServer(database *accresults.Database) error {
    accdb = database
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/event/", eventHandler)
    http.HandleFunc("/player/", playerHandler)
    http.HandleFunc("/session/", sessionHandler)
    http.HandleFunc("/sessioncar/", sessionCarHandler)
    return http.ListenAndServe(":8099", nil)
}
