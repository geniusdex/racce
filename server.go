package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
    "github.com/geniusdex/racce/accresults"
)

var (
    session *accresults.Session
)

func sessionHandler(w http.ResponseWriter, r *http.Request) {
    functions := template.FuncMap{
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
    
    t,err := template.New("session.html").Funcs(functions).ParseFiles("session.html")
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

func RunServer(inSession *accresults.Session) error {
    session = inSession
    http.HandleFunc("/", sessionHandler)
    return http.ListenAndServe(":8099", nil)
}
