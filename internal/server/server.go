package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog"

	"Exlord/otpservice/internal/services"
	"Exlord/otpservice/internal/storage"
)

func Start(addr, jwtSecret string) error {
	logger := httplog.NewLogger("otpservice", httplog.Options{JSON: true})
	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// serve openapi
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.ReadFile("openapi.yaml")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(200)
		_, _ = w.Write(f)
	})
	r.Get("/openapi", func(w http.ResponseWriter, r *http.Request) {
		h := `<!doctype html><html><head><title>OTP API</title></head>
		<body style="margin:0;font-family:system-ui">
		<div style="padding:16px"><h1>OTP Service API</h1>
		<p>Open <a href="/openapi.yaml">openapi.yaml</a> in <a href="https://editor.swagger.io/" target="_blank">Swagger Editor</a>.</p></div>
		</body></html>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(h))
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})

	// dependencies
	userRepo := storage.NewInMemoryUserRepo()
	otpStore := storage.NewInMemoryOTPStore()
	ratelimiter := services.NewRateLimiter(3, 10*time.Minute)
	auth := services.NewAuthService(jwtSecret)
	otpSvc := services.NewOTPService(otpStore, ratelimiter)
	userSvc := services.NewUserService(userRepo)

	// routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/request-otp", func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Phone string `json:"phone"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Phone) == "" {
					http.Error(w, "invalid body", 400)
					return
				}
				code, err := otpSvc.Generate(r.Context(), body.Phone)
				if err != nil {
					if errorsIsRate(err) {
						http.Error(w, err.Error(), 429)
						return
					}
					http.Error(w, err.Error(), 400)
					return
				}
				log.Printf("======================================")
				log.Printf("[OTP] phone=%s code=%s", body.Phone, code) // print to console
				log.Printf("======================================")
				respondJSON(w, 200, map[string]string{"message": "otp generated"})
			})
			r.Post("/verify", func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Phone string `json:"phone"`
					OTP   string `json:"otp"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Phone) == "" || strings.TrimSpace(body.OTP) == "" {
					http.Error(w, "invalid body", 400)
					return
				}
				ok := otpSvc.Verify(r.Context(), body.Phone, body.OTP)
				if !ok {
					http.Error(w, "invalid or expired otp", 400)
					return
				}
				// upsert user
				u, err := userSvc.UpsertByPhone(r.Context(), body.Phone)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				tok, err := auth.IssueJWT(u.ID.String())
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				respondJSON(w, 200, map[string]any{"token": tok, "user": u})
			})
		})
		// protected
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware)
			r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
				search := strings.TrimSpace(r.URL.Query().Get("search"))
				page := atoiDefault(r.URL.Query().Get("page"), 1)
				pageSize := atoiDefault(r.URL.Query().Get("pageSize"), 10)
				items, total := userSvc.List(r.Context(), search, page, pageSize)
				respondJSON(w, 200, map[string]any{"items": items, "page": page, "pageSize": pageSize, "total": total})
			})
			r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				u, ok := userSvc.Get(r.Context(), id)
				if !ok {
					http.NotFound(w, r)
					return
				}
				respondJSON(w, 200, u)
			})
		})
	})
	return http.ListenAndServe(addr, r)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	i, err := strconv.Atoi(s)
	if err != nil || i <= 0 {
		return def
	}
	return i
}

func errorsIsRate(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "rate")
}
