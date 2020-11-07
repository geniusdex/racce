package frontend

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/geniusdex/racce/accresults"
)

func (f *frontend) indexHandler(w http.ResponseWriter, r *http.Request) {
	if len(strings.Trim(r.URL.Path, "/")) > 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.db.Mutex.RLock()
	defer f.db.Mutex.RUnlock()

	f.executeTemplate(w, r, "index.html", f.db)
}

func (f *frontend) eventHandler(w http.ResponseWriter, r *http.Request) {
	pathComponents := strings.Split(r.URL.Path, "/")
	if len(pathComponents) != 3 {
		log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.db.Mutex.RLock()
	defer f.db.Mutex.RUnlock()

	eventID := pathComponents[2]
	event, ok := f.db.Events[eventID]
	if !ok {
		log.Printf("Event '%s' not found in database", eventID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.executeTemplate(w, r, "event.html", event)
}

type playerPage struct {
	Database *accresults.Database
	Player   *accresults.Player
}

func (f *frontend) playerHandler(w http.ResponseWriter, r *http.Request) {
	pathComponents := strings.Split(r.URL.Path, "/")
	if len(pathComponents) != 3 {
		log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.db.Mutex.RLock()
	defer f.db.Mutex.RUnlock()

	playerID := pathComponents[2]
	player, ok := f.db.Players[playerID]
	if !ok {
		log.Printf("Player '%s' not found in database", playerID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.executeTemplate(w, r, "player.html", &playerPage{f.db, player})
}

func (f *frontend) sessionHandler(w http.ResponseWriter, r *http.Request) {
	pathComponents := strings.Split(r.URL.Path, "/")
	if len(pathComponents) != 3 {
		log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.db.Mutex.RLock()
	defer f.db.Mutex.RUnlock()

	sessionName := pathComponents[2]
	session, ok := f.db.Sessions[sessionName]
	if !ok {
		log.Printf("Session '%s' not found in database", sessionName)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.executeTemplate(w, r, "session.html", session)
}

type sessionCarPage struct {
	Session *accresults.Session
	Car     *accresults.Car
}

func (f *frontend) sessionCarHandler(w http.ResponseWriter, r *http.Request) {
	pathComponents := strings.Split(r.URL.Path, "/")
	if len(pathComponents) != 4 {
		log.Printf("Not enough components in path '%s', components: %v, len: %v", r.URL.Path, pathComponents, len(pathComponents))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.db.Mutex.RLock()
	defer f.db.Mutex.RUnlock()

	sessionName := pathComponents[2]
	session, ok := f.db.Sessions[sessionName]
	if !ok {
		log.Printf("Session '%s' not found in database", sessionName)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	carID, err := strconv.Atoi(pathComponents[3])
	if err != nil {
		log.Printf("Car ID '%s' is not numeric", pathComponents[3])
		w.WriteHeader(http.StatusNotFound)
		return
	}
	car := session.FindCarById(carID)
	if car == nil {
		log.Printf("Car ID '%d' not present in session", carID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f.executeTemplate(w, r, "sessioncar.html", &sessionCarPage{session, car})
}
