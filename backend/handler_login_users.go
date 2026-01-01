package main

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

type RefershTokenParams struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

func (cfg *apiConfig) handlerLoginUsers(w http.ResponseWriter, r *http.Request) {
	type response struct {
		User
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "bad request", err)
	}

	email := params.Email
	_, err = mail.ParseAddress(params.Email)
	if err != nil {
		respondWithError(w, 400, "please provide a valid email", err)
		return
	}

}
