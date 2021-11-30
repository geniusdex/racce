package frontend

import (
	"fmt"
	"html/template"
	"image/color"
	"log"
	"net/http"

	"github.com/geniusdex/racce/accdata"
	"github.com/geniusdex/racce/accresults"
	"github.com/geniusdex/racce/accserver"
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
	// AdminWithoutPassword allows logging in without specifying a password at all. Use
	// with care! This is primarily meant for development.
	AdminWithoutPassword bool `json:"adminWithoutPassword"`
	// Live indicates the live server status page should be enabled.
	Live bool `json:"live"`
	// DisableTemplateCache will parse templates on every page load, instead of at startup
	DisableTemplateCache bool `json:"disableTemplateCache"`
}

// templateData contains dynamic data to be used while rendering templates
type templateData struct {
	basePath string
}

type frontend struct {
	config       *Configuration
	db           *accresults.Database
	server       *accserver.Server
	templates    *template.Template
	templateData *templateData
}

func (f *frontend) addTemplateFunctions(t *template.Template, templateData *templateData) *template.Template {
	t.Funcs(template.FuncMap(qogs.TemplateFuncs()))
	t.Funcs(template.FuncMap{
		// Environment
		"basePath": func() string {
			return templateData.basePath
		},
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
		"color": func(color color.RGBA) template.CSS {
			return template.CSS(fmt.Sprintf("rgba(%d, %d, %d, %g)", color.R, color.G, color.B, float32(color.A)/255.0))
		},
		// Data specific for Assetto Corsa Competizione
		"track": func(name string) *accdata.Track {
			if track := accdata.TrackByLabel(name); track != nil {
				return track
			}
			return &accdata.Track{"-", "-", accdata.Competition{"-"}, 0, 0, []string{}}
		},
		"tracks": func() []*accdata.Track {
			return accdata.Tracks
		},
		"carmodel": func(id int) *accdata.CarModel {
			if model := accdata.CarModelByID(id); model != nil {
				return model
			}
			return &accdata.CarModel{0, "-", "-", "-", 0, accdata.GT3}
		},
		"carmodels": func() []*accdata.CarModel {
			return accdata.CarModels
		},
		"drivercategory": func(id int) *accdata.DriverCategory {
			if category := accdata.DriverCategoryByID(id); category != nil {
				return category
			}
			return &accdata.DriverCategory{0, "-"}
		},
		"cupcategories": func() []*accdata.DriverCategory {
			return accdata.DriverCategories
		},
		"cupcategory": func(id int) *accdata.CupCategory {
			if category := accdata.CupCategoryByID(id); category != nil {
				return category
			}
			return &accdata.CupCategory{0, "-", color.RGBA{}, color.RGBA{}}
		},
		"drivercategories": func() []*accdata.CupCategory {
			return accdata.CupCategories
		},
		// Information about racce instance
		"isLiveStateEnabled": func() bool {
			return f.config.Live
		},
	})
	return t
}

func (f *frontend) initializeTemplates() (*template.Template, error) {
	return f.addTemplateFunctions(template.New("templates"), f.templateData).ParseGlob("templates/*.html")
}

func (f *frontend) getTemplates() (*template.Template, error) {
	if f.templates != nil {
		return f.templates, nil
	}
	return f.initializeTemplates()
}

func basePath(r *http.Request) string {
	return r.Header.Get("X-Forwarded-Prefix")
}

func (f *frontend) executeTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	sessions.Save(r, w)

	f.templateData.basePath = basePath(r)

	t, err := f.getTemplates()
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
func Run(config *Configuration, database *accresults.Database, server *accserver.Server) error {
	if config.Live && server == nil {
		log.Printf("Live server state in frontend is enabled in configuration, but server is not managed; disabling live page")
		config.Live = false
	}

	f := &frontend{
		config,
		database,
		server,
		nil,
		&templateData{},
	}

	if !config.DisableTemplateCache {
		templates, err := f.initializeTemplates()
		if err != nil {
			return fmt.Errorf("cannot initialize templates: %w", err)
		}
		f.templates = templates
	}

	http.HandleFunc("/", f.indexHandler)
	http.HandleFunc("/indexfull", f.indexFullHandler)
	http.HandleFunc("/indexfull/", f.indexFullHandler)
	http.HandleFunc("/event/", f.eventHandler)
	http.HandleFunc("/player/", f.playerHandler)
	http.HandleFunc("/playertrack/", f.playerTrackHandler)
	http.HandleFunc("/session/", f.sessionHandler)
	http.HandleFunc("/sessioncar/", f.sessionCarHandler)
	if config.Live {
		http.HandleFunc("/live/", f.liveHandler)
		http.HandleFunc("/live/ws", f.liveWebSocketHandler)
	}

	admin := newAdmin(config, server, f)
	http.Handle("/admin/", admin)
	http.Handle("/admin", admin)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return http.ListenAndServe(config.Listen, nil)
}
