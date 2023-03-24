package gateway

import (
	"net/http"

	"github.com/go-kit/log/level"
)

func (conf *Configuration) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := false
		if conf.AuthType == "basic" {
			ok = conf.basicAuth(w, r)
		}

		if !ok {
			level.Debug(conf.Logger).Log("msg", "No valid tenant credentials are found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (conf *Configuration) basicAuth(w http.ResponseWriter, r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	validTenant := false
	for _, tenant := range conf.Tenants {
		if tenant.Username == username {
			if tenant.Password == password {
				r.Header.Set("X-Scope-OrgID", tenant.XScopeOrgId)
				validTenant = true
				break
			} else {
				level.Debug(conf.Logger).Log("msg", "Wrong password for ", tenant.Username)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return false
			}
		}
	}
	return validTenant
}
