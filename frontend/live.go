package frontend

import (
	"net/http"

	"github.com/geniusdex/racce/accserver"
)

type livePage struct {
	Server *accserver.Server
}

func (f *frontend) liveHandler(w http.ResponseWriter, r *http.Request) {
	page := &livePage{
		Server: f.server,
	}

	f.executeTemplate(w, r, "live.html", page)
}
