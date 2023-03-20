package gateway

import (
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func (tenants *Tenant) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.NewLogfmtLogger(os.Stdout)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		noValidTenant := true
		for key, value := range tenants.All {
			if value.Username == username {
				if value.Password == password {
					r.Header.Set("X-Scope-OrgID", value.ID)
					noValidTenant = false
				} else {
					level.Error(logger).Log("msg", "Wrong password for ", key)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
		}

		if noValidTenant {
			level.Error(logger).Log("msg", "No valid tenant credentials are found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
