package gateway

import (
	"crypto/subtle"
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
			logrus.Debugf("No valid tenant credentials are found")
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

	if !tenant.saveCompare(username, password) {
		return false
	}

	if !tenant.Passthrough {
		r.Header.Set("X-Scope-OrgID", tenant.ID)
	}
	return true
}

// attempt to mitigate timing attacks
func (tenant *Tenant) saveCompare(username, password string) bool {
	userNameCheck := subtle.ConstantTimeCompare([]byte(tenant.Username), []byte(username))
	passwordCheck := subtle.ConstantTimeCompare([]byte(tenant.Password), []byte(password))
	if userNameCheck == 1 && passwordCheck == 1 {
		return true
	}
	return false
}
