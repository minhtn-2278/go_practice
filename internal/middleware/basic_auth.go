package middleware

import (
	"net/http"
)

type User struct {
	user     string
	password string
}

var users = []User{
	{user: "admin", password: "password1"},
	{user: "user", password: "password2"},
}

func BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providedUsername, providedPassword, ok := r.BasicAuth()
		if !ok || !isValidUser(providedUsername, providedPassword) {
			w.Header().Set("WWW-Authenticate", `Basic realm="employee-management"`)
			writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isValidUser(username, password string) bool {
	for _, currentUser := range users {
		if username == currentUser.user && password == currentUser.password {
			return true
		}
	}

	return false
}
