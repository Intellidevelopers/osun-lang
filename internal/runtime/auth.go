package runtime

import (
	"fmt"
	"net/http"
)

func RequireAuth(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	fmt.Println("🔑 Auth token:", token)
}
