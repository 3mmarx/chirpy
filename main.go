package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())

	w.Write([]byte(msg))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	cfg.fileserverHits.Store(0)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))
}

func validatedChripHandler(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Body string `json:"body"`
	}

	type errResp struct {
		Error string `json:"error"`
	}

	type validResp struct {
		CleanedBody string `json:"cleaned_body"`
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	var reqBody req
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp{Error: "Something went wrong"})
		return
	}

	if len(reqBody.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp{Error: "Chirp is too long"})
		return
	}

	forbiddenWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Fields(reqBody.Body)
	for i, word := range words {
		if _, ok := forbiddenWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}

	cleanedBody := strings.Join(words, " ")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(validResp{CleanedBody: cleanedBody})
}

func main() {
	mux := http.NewServeMux()

	apiCfg := apiConfig{}

	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validatedChripHandler)

	s := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Fatal(s.ListenAndServe())
}
