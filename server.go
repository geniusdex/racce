package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
    "reflect"
    "sort"
    "strings"
    "strconv"
    "github.com/geniusdex/racce/accresults"
)

var (
    accdb *accresults.Database
)

func templateFunctionMap(basePath string) template.FuncMap{
    return template.FuncMap{
    // Arithmetic
        "add": func(a, b int) int {
            return a + b
        },
        "sub": func(a, b int) int {
            return a + b
        },
        "div": func(a, b int) float64 {
            return float64(a) / float64(b)
        },
    // Formatting
        "laptime": func(time int) string {
            milliseconds := time % 1000
            time = time / 1000
            seconds := time % 60
            minutes := time / 60
            
            return fmt.Sprintf("%d:%02d.%03d", minutes, seconds, milliseconds)
        },
    // Utility
        "contains": func(haystack []string, needle string) bool {
            for _, element := range haystack {
                if element == needle {
                    return true
                }
            }
            return false
        },
        "keys": func(data interface{}) []string {
            value := reflect.ValueOf(data)
            keys := make([]string, value.Len())
            for i, key := range value.MapKeys() {
                keys[i] = key.String()
            }
            return keys
        },
        "sort": func(in []string) []string {
            out := make([]string, len(in))
            copy(out, in)
            sort.Strings(out)
            return out
        },
        "sortOn": func(data interface{}, field string) []interface{} {
            dataval := reflect.ValueOf(data)
            values := make([]interface{}, 0, dataval.Len())
            if dataval.Kind() == reflect.Map {
                iter := dataval.MapRange()
                for iter.Next() {
                    values = append(values, iter.Value().Elem().Interface())
                }
            } else {
                for i := 0; i < dataval.Len(); i++ {
                    values = append(values, dataval.Index(i).Elem().Interface())
                }
            }
            sort.Slice(values, func(i, j int) bool {
                a := reflect.ValueOf(values[i]).FieldByName(field)
                b := reflect.ValueOf(values[j]).FieldByName(field)
                return a.String() < b.String()
            })
            return values
        },
        "reverse": func(data interface{}) []interface{} {
            dataval := reflect.ValueOf(data)
            values := make([]interface{}, dataval.Len())
            for i, o := dataval.Len() - 1, 0; i >= 0; i, o = i-1, o+1 {
                values[o] = dataval.Index(i).Interface()
            }
            return values
        },
    // Environment
        "basePath": func() string {
            return basePath
        },
    }
}

func executeTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
    basePath := r.Header.Get("X-Forwarded-Prefix")
    
    t,err := template.New("templates").Funcs(templateFunctionMap(basePath)).ParseGlob("templates/*.html")
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
    if len(strings.Trim(r.URL.Path, "/")) > 0 {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    accdb.Mutex.RLock()
    defer accdb.Mutex.RUnlock()
    
    executeTemplate(w, r, "index.html", accdb)
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
    
    executeTemplate(w, r, "event.html", event)
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
    
    executeTemplate(w, r, "player.html", &playerPage{accdb, player})
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
    
    executeTemplate(w, r, "session.html", session)
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
    
    executeTemplate(w, r, "sessioncar.html", &sessionCarPage{session, car})
}

func RunServer(database *accresults.Database) error {
    accdb = database
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/event/", eventHandler)
    http.HandleFunc("/player/", playerHandler)
    http.HandleFunc("/session/", sessionHandler)
    http.HandleFunc("/sessioncar/", sessionCarHandler)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    return http.ListenAndServe(":8099", nil)
}
