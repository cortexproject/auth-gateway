package gateway

import (
	"log"
	"net/http"
)

func Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := authenticateAllTenants(w, r)
		if !ok {
			log.Println("Not all tenants have a valid credentials")
			return
		}
		h.ServeHTTP(w, r)
	})
}

func authenticateSingleTenant(w http.ResponseWriter, r *http.Request, t *Tenant) bool {
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

	if username != t.Username || password != t.Password {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	r.Header.Set("X-Scope-OrgID", t.ID)

	return true
}

func authenticateAllTenants(w http.ResponseWriter, r *http.Request) bool {
	tenants := GetTenants()
	for _, tenant := range tenants {
		ok := authenticateSingleTenant(w, r, &tenant)
		if !ok {
			log.Printf("Could not authenticate this tenant: %v\n", tenant)
			return false
		}
	}

	return true
}
