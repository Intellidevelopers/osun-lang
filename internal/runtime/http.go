package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type OsunServer struct {
	port        int
	middlewares []Middleware
	routes      map[string]map[string]http.HandlerFunc
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

// Create a new Osun HTTP server
func NewOsunServer(port int) *OsunServer {
	return &OsunServer{
		port:   port,
		routes: make(map[string]map[string]http.HandlerFunc),
	}
}

// Add middleware
func (s *OsunServer) Use(m Middleware) {
	s.middlewares = append(s.middlewares, m)
}

// Register route by method
func (s *OsunServer) Handle(method, path string, handler http.HandlerFunc) {
	method = strings.ToUpper(method)
	if s.routes[method] == nil {
		s.routes[method] = make(map[string]http.HandlerFunc)
	}
	s.routes[method][path] = handler
}

// Start server
func (s *OsunServer) Listen() {
	for method, paths := range s.routes {
		for path, handler := range paths {
			h := handler
			for i := len(s.middlewares) - 1; i >= 0; i-- {
				h = s.middlewares[i](h)
			}
			http.HandleFunc(path, methodHandler(method, h))
			fmt.Printf("[%s] %s registered\n", method, path)
		}
	}

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("🚀 Osun Server running at http://localhost%s\n", addr)
	http.ListenAndServe(addr, nil)
}

func methodHandler(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

// --- Helpers ---
func ParseJSON(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func WriteText(w http.ResponseWriter, status int, text string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(text))
}

// ---------------- JWT ----------------

func GenerateToken(userID string) (string, error) {
	secret := GetEnv("JWT_SECRET", "osun_secret")
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string) (string, error) {
	secret := GetEnv("JWT_SECRET", "osun_secret")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	userID, _ := claims["user_id"].(string)
	return userID, nil
}
