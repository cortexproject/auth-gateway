package gateway

import (
	"net/http"

	"github.com/go-kit/log/level"
)

type Authentication struct {
	config *Config
}

func NewAuthentication(config *Config) *Authentication {
	return &Authentication{
		config: config,
	}
}

func (a Authentication) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := false
		for _, tenant := range a.config.Tenants {
			if tenant.Authentication == "basic" {
				ok = tenant.basicAuth(w, r)
				if ok {
					break
				}
			}
			// add other authentication methods if necessary
		}
		if !ok {
			level.Debug(logger).Log("msg", "No valid tenant credentials are found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (tenant *Tenant) basicAuth(w http.ResponseWriter, r *http.Request) bool {
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

	if tenant.Username == username {
		if tenant.Password == password {
			r.Header.Set("X-Scope-OrgID", tenant.ID)
			return true
		} else {
			level.Debug(logger).Log("msg", "Wrong password for ", tenant.Username)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return false
		}
	}

	return false
}
