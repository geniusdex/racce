package frontend

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

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

type frontend struct {
	config *Configuration
	db     *accresults.Database
}

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

// Run runs the frontend with the given configuration and database
func Run(config *Configuration, database *accresults.Database) error {
	f := &frontend{
		config,
		database,
	}

	http.HandleFunc("/", f.indexHandler)
	http.HandleFunc("/event/", f.eventHandler)
	http.HandleFunc("/player/", f.playerHandler)
	http.HandleFunc("/session/", f.sessionHandler)
	http.HandleFunc("/sessioncar/", f.sessionCarHandler)

	admin := newAdmin(config)
	http.Handle("/admin/", admin)
	http.Handle("/admin", admin)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return http.ListenAndServe(config.Listen, nil)
}
