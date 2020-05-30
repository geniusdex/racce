package frontend

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/geniusdex/racce/accresults"
	"github.com/geniusdex/racce/qogs"
	"github.com/gorilla/sessions"
)

// Configuration specifies the frontend configuration
type Configuration struct {
	// Listen specifies the IP and port to listen on. IP can be empty for all interfaces.
	Listen string `json:"listen"`
	// AdminPassword specifies the password needed to login as an admin. Leave empty to
	// disable admin access.
	AdminPassword string `json:"adminPassword"`
}

var (
	accdb *accresults.Database
)

func addTemplateFunctions(t *template.Template, basePath string) *template.Template {
	t.Funcs(template.FuncMap(qogs.TemplateFuncs()))
	t.Funcs(template.FuncMap{
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
		"mul": func(a, b int) int {
			return a * b
		},
		// Formatting
		"laptime": func(time int) string {
			milliseconds := time % 1000
			time = time / 1000
			seconds := time % 60
			minutes := time / 60

			return fmt.Sprintf("%d:%02d.%03d", minutes, seconds, milliseconds)
		},
		// Environment
		"basePath": func() string {
			return basePath
		},
	})
	return t
}

func basePath(r *http.Request) string {
	return r.Header.Get("X-Forwarded-Prefix")
}

func executeTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	sessions.Save(r, w)

	t, err := addTemplateFunctions(template.New("templates"), basePath(r)).ParseGlob("templates/*.html")
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
	Player   *accresults.Player
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
	Car     *accresults.Car
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
	carId, err := strconv.Atoi(pathComponents[3])
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

// Run runs the frontend with the given configuration and database
func Run(config *Configuration, database *accresults.Database) error {
	accdb = database
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/event/", eventHandler)
	http.HandleFunc("/player/", playerHandler)
	http.HandleFunc("/session/", sessionHandler)
	http.HandleFunc("/sessioncar/", sessionCarHandler)

	admin := newAdmin(config)
	http.Handle("/admin/", admin)
	http.Handle("/admin", admin)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return http.ListenAndServe(config.Listen, nil)
}
