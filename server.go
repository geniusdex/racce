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

func templateFunctionMap(session *accresults.Session) template.FuncMap{
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
        "car": func(carId int) *accresults.Car {
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
    t,err := template.New("index.html").Funcs(templateFunctionMap(nil)).ParseFiles("index.html")
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
    
    err = t.Execute(w, accdb)
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
    pathComponents := strings.Split(r.URL.Path, "/")
    if len(pathComponents) != 3 {
        log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    eventId := pathComponents[2]
    event, ok := accdb.Events[eventId]
    if !ok {
        log.Printf("Event '%s' not found in database", eventId)
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    t,err := template.New("event.html").Funcs(templateFunctionMap(nil)).ParseFiles("event.html")
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
    
    err = t.Execute(w, event)
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
}
func sessionHandler(w http.ResponseWriter, r *http.Request) {
    pathComponents := strings.Split(r.URL.Path, "/")
    if len(pathComponents) != 3 {
        log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    sessionName := pathComponents[2]
    session, ok := accdb.Sessions[sessionName]
    if !ok {
        log.Printf("Session '%s' not found in database", sessionName)
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    t,err := template.New("session.html").Funcs(templateFunctionMap(session)).ParseFiles("session.html")
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
    
    err = t.Execute(w, session)
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
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
    
    t,err := template.New("sessioncar.html").Funcs(templateFunctionMap(session)).ParseFiles("sessioncar.html")
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
    
    err = t.Execute(w, &sessionCarPage{session, car})
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
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
    
    playerId := pathComponents[2]
    player, ok := accdb.Players[playerId]
    if !ok {
        log.Printf("Player '%s' not found in database", playerId)
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    t,err := template.New("player.html").Funcs(templateFunctionMap(nil)).ParseFiles("player.html")
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
    
    err = t.Execute(w, &playerPage{accdb, player})
    if err != nil {
        log.Print(err)
        w.Write([]byte(err.Error()))
    }
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
