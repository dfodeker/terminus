package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type StoreResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Handle    string    `json:"handle"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cfg *apiConfig) handlerCreateStore(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		name   string `json:"name"`
		handle string `json:"handle"`
		plan   string `json:"plan"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error Decoding Params: %s", err)
		respondWithError(w, 400, "Please provide a valid request body", err)
		return
	}
}
