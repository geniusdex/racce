package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
    "strings"
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

func RunServer(database *accresults.Database) error {
    accdb = database
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/session/", sessionHandler)
    return http.ListenAndServe(":8099", nil)
}
