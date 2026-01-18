package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/agent"
	"github.com/achetronic/magec/server/config"
	"github.com/achetronic/magec/server/logging"
)

var configFile = flag.String("config", "config.yaml", "Path to config file")

func main() {
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logging.Setup(cfg.Log.Level, cfg.Log.Format)

	ctx := context.Background()

	agentService, err := agent.New(ctx, cfg)
	if err != nil {
		slog.Error("Failed to initialize agent", "error", err)
		os.Exit(1)
	}
	slog.Info("Agent initialized")

	mux := http.NewServeMux()

	// All API routes under /api/v1/
	mux.Handle("/api/v1/agent/", http.StripPrefix("/api/v1/agent", agentService.Handler()))

	if cfg.Transcription.Resolved != nil {
		if target, err := url.Parse(cfg.Transcription.Resolved.URL); err == nil {
			mux.Handle("/api/v1/transcription/", newTranscriptionProxy(target))
			slog.Debug("Transcription proxy enabled", "target", target.String())
		}
	}

	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// TTS proxy
	if cfg.TTS.Resolved != nil {
		if target, err := url.Parse(cfg.TTS.Resolved.URL); err == nil {
			mux.Handle("/api/v1/tts/", newTTSProxy(target, &cfg.TTS))
			slog.Debug("TTS proxy enabled", "target", target.String())
		}
	}

	// Log registered routes
	logRoutes(agentService.Handler())

	// Static files
	mux.Handle("/", http.FileServer(http.Dir("gui")))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      accessLogMiddleware(corsMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	slog.Info("Server started", "addr", addr, "url", fmt.Sprintf("http://%s", addr))
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

// logRoutes logs all registered API routes at debug level.
func logRoutes(adkHandler http.Handler) {
	slog.Debug("Registered routes:")

	// ADK routes
	if router, ok := adkHandler.(*mux.Router); ok {
		router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
			if path, err := route.GetPathTemplate(); err == nil {
				methods, _ := route.GetMethods()
				slog.Debug("Route", "methods", methods, "path", "/api/v1/agent"+path)
			}
			return nil
		})
	}

	slog.Debug("Route", "methods", []string{"POST"}, "path", "/api/v1/transcription/")
	slog.Debug("Route", "methods", []string{"POST"}, "path", "/api/v1/tts/")
	slog.Debug("Route", "methods", []string{"GET"}, "path", "/api/v1/health")
	slog.Debug("Route", "methods", []string{"GET"}, "path", "/ (static files)")
}

// responseRecorder wraps http.ResponseWriter to capture status code and bytes written.
type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// accessLogMiddleware logs all HTTP requests with method, path, status, duration and size.
func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		logFn := slog.Info
		if rec.status >= 400 {
			logFn = slog.Warn
		}
		logFn("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", duration.Round(time.Millisecond),
			"bytes", rec.bytes,
		)
	})
}

// corsMiddleware adds CORS headers to all responses.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// newTranscriptionProxy creates a reverse proxy for the transcription API.
func newTranscriptionProxy(target *url.URL) http.Handler {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			// /api/v1/transcription/audio/transcriptions -> /v1/audio/transcriptions
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1/transcription")
			req.Host = target.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("Transcription proxy error", "error", err, "path", r.URL.Path)
			http.Error(w, "Transcription service unavailable", http.StatusBadGateway)
		},
	}
}

// newTTSProxy creates a reverse proxy for the TTS API (OpenAI-compatible).
// It injects the configured model, voice, speed, and format into the request body.
func newTTSProxy(target *url.URL, ttsCfg *config.TTSConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read original body
		var inputBody map[string]interface{}
		if r.Body != nil {
			body, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			if len(body) > 0 {
				if err := json.Unmarshal(body, &inputBody); err != nil {
					http.Error(w, "Invalid JSON body", http.StatusBadRequest)
					return
				}
			}
		}
		if inputBody == nil {
			inputBody = make(map[string]interface{})
		}

		// Inject config values (server config takes precedence)
		inputBody["model"] = ttsCfg.Model
		inputBody["voice"] = ttsCfg.Voice
		inputBody["speed"] = ttsCfg.Speed

		// Marshal new body
		newBody, err := json.Marshal(inputBody)
		if err != nil {
			http.Error(w, "Failed to build request", http.StatusInternalServerError)
			return
		}

		// Create proxied request
		proxyURL := *target
		proxyURL.Path = "/v1/audio/speech"
		
		proxyReq, err := http.NewRequestWithContext(r.Context(), "POST", proxyURL.String(), bytes.NewReader(newBody))
		if err != nil {
			http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
			return
		}
		proxyReq.Header.Set("Content-Type", "application/json")
		if ttsCfg.Resolved != nil && ttsCfg.Resolved.APIKey != "" {
			proxyReq.Header.Set("Authorization", "Bearer "+ttsCfg.Resolved.APIKey)
		}

		// Forward request
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(proxyReq)
		if err != nil {
			slog.Error("TTS proxy error", "error", err)
			http.Error(w, "TTS service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers and body
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
}
