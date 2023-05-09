package gateway

import (
	"net/http"

	"github.com/cortexproject/auth-gateway/middleware"
	"github.com/sirupsen/logrus"
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
		sr := &middleware.StatusRecorder{
			ResponseWriter: w,
		}
		ok := false
		for _, tenant := range a.config.Tenants {
			if tenant.Authentication == "basic" {
				ok = tenant.basicAuth(sr, r)
				if ok {
					break
				}
			}
			// add other authentication methods if necessary
		}

		if ok {
			next.ServeHTTP(sr, r)
		} else {
			logrus.Infof("No valid tenant credentials are found")
			sr.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(sr, "Unauthorized", http.StatusUnauthorized)
		}
	})
}

func (tenant *Tenant) basicAuth(w http.ResponseWriter, r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}
	username, password, ok := r.BasicAuth()
	if !ok {
		return false
	}

	if tenant.Username == username {
		if tenant.Password == password {
			r.Header.Set("X-Scope-OrgID", tenant.ID)
			return true
		} else {
			return false
		}
	}

	return false
}
