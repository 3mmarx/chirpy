package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func validateChripHandler(w http.ResponseWriter, r *http.Request) {
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
