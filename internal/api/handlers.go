package api

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"portfolio-server/internal/data"
	"portfolio-server/internal/discord"
	"portfolio-server/internal/renderer"
	"portfolio-server/internal/sshbox"
)

// ─────────────────────────────────────────────
//  REGISTER ROUTES
// ─────────────────────────────────────────────

func RegisterRoutes(mux *http.ServeMux, store *data.Store, dc *discord.Client) {
	// Static files: CSS, JS, Images
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Root → serves the dynamic template
	mux.HandleFunc("/", serveIndex(store))

	// API
	mux.HandleFunc("/api/contact", handleContact(dc))
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/cmd", handleTerminalCmd(store, dc))
}

// ─────────────────────────────────────────────
//  HANDLERS
// ─────────────────────────────────────────────

func serveIndex(store *data.Store) http.HandlerFunc {
	funcs := template.FuncMap{
		"split": strings.Split,
		"add":   func(a, b int) int { return a + b },
		"title": strings.Title,
		"cleanNum": func(v string) string {
			for i, c := range v {
				if c < '0' || c > '9' {
					return v[:i]
				}
			}
			return v
		},
		"env": func(key string) string {
			return os.Getenv(key)
		},
		"mod": func(a, b int) int { return a % b },
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"asciiProfile": func() template.HTML {
			var imgData []byte
			b64 := os.Getenv("PROFILE_IMAGE_BASE64")
			if b64 != "" {
				imgData, _ = base64.StdEncoding.DecodeString(b64)
			}

			var html string
			var err error
			// Slightly smaller width for better web fit (50 chars)
			if len(imgData) > 0 {
				html, err = renderer.ImageBytesToHTMLHalfBlock(imgData, 50)
			} else {
				// Fallback to local file if no env var
				html, err = renderer.ImageToHTMLHalfBlock("assets/profile.jpg", 50)
			}

			if err != nil {
				log.Printf("[http] ascii profile error: %v", err)
				return template.HTML("<!-- ASCII rendering failed -->")
			}
			return template.HTML(html)
		},
	}
	tmpl := template.Must(template.New("index.html").Funcs(funcs).ParseFiles("tmpl/index.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, store); err != nil {
			log.Printf("[http] template error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

type contactRequest struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type apiResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func handleContact(d *discord.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req contactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(apiResponse{OK: false, Message: "invalid JSON"})
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		req.Message = strings.TrimSpace(req.Message)

		if req.Name == "" || req.Message == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(apiResponse{OK: false, Message: "name and message are required"})
			return
		}

		log.Printf("[http] contact from %q: %s", req.Name, req.Message)
		d.Send(req.Message, req.Name+" (Web Form)")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(apiResponse{OK: true, Message: "Message sent! Puneet will get back to you."})
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"server": "portfolio-v1",
	})
}

func handleTerminalCmd(store *data.Store, dc *discord.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Command string `json:"command"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(apiResponse{OK: false, Message: "invalid JSON"})
			return
		}

		output := sshbox.ProcessCommand(req.Command, store, dc)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"output": output,
		})
	}
}
