package frontend

import (
	"log"
	"net/http"

	"github.com/geniusdex/racce/accserver"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

type admin struct {
	config   *Configuration
	store    sessions.Store
	serveMux *http.ServeMux
	server   *accserver.Server
}

func newAdmin(config *Configuration, accServer *accserver.Server) *admin {
	admin := &admin{
		config,
		sessions.NewCookieStore(securecookie.GenerateRandomKey(32)),
		http.NewServeMux(),
		accServer,
	}

	admin.serveMux.HandleFunc("/admin/server", admin.serverHandler)
	admin.serveMux.HandleFunc("/admin", admin.indexHandler)

	return admin
}

type adminLoginPage struct {
	InvalidPassword bool
}

func (a *admin) loginHandler(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	page := &adminLoginPage{false}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			log.Panicf("Cannot parse admin login form: %v", err)
		}

		if a.config.AdminPassword == "" {
			log.Printf("Admin login disabled")
		} else if r.FormValue("password") != a.config.AdminPassword {
			log.Printf("Admin login failed: invalid password")
		} else {
			session.Values["loggedIn"] = true
			session.Save(r, w)
			http.Redirect(w, r, basePath(r)+"/admin/", http.StatusSeeOther)
			return
		}

		page.InvalidPassword = true
	}

	executeTemplate(w, r, "admin-login.html", page)
}

type adminIndexPage struct {
	Server *accserver.Server
}

func (a *admin) indexHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, r, "admin.html", &adminIndexPage{a.server})
}

func (a *admin) serverHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, r, "admin-server.html", a.server)
}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionName := "admin-session"
	session, err := a.store.Get(r, sessionName)
	if err != nil {
		log.Printf("Could not get existing admin session: %v", err)
	}

	if r.URL.Path == "/admin/login" {
		a.loginHandler(w, r, session)
		return
	}

	if !a.config.AdminWithoutPassword {
		if loggedIn, ok := session.Values["loggedIn"]; !ok || !loggedIn.(bool) {
			http.Redirect(w, r, basePath(r)+"/admin/login", http.StatusSeeOther)
		}
	}

	a.serveMux.ServeHTTP(w, r)
}
