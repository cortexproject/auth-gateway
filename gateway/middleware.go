package gateway

import (
	"net/http"

	"github.com/go-kit/log/level"
)

func (tenants *Tenant) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		validTenant := false
		for key, value := range tenants.All {
			if value.Username == username {
				if value.Password == password {
					r.Header.Set("X-Scope-OrgID", value.ID)
					validTenant = true
					break
				} else {
					level.Debug(tenants.Logger).Log("msg", "Wrong password for ", key)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
		}

		if !validTenant {
			level.Debug(tenants.Logger).Log("msg", "No valid tenant credentials are found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
