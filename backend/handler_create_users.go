package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/dfodeker/terminus/internal/auth"
	"github.com/dfodeker/terminus/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error Decoding Params: %s", err)
		respondWithError(w, 400, "Please provide a valid request body", err)
		return
	}
	email := params.Email
	_, err = mail.ParseAddress(params.Email)
	if err != nil {
		respondWithError(w, 400, "Please Provide a Valid Email", err)
		return
	}
	pass := params.Password
	hash, err := auth.HashPassword(pass)
	if err != nil {
		respondWithError(w, 500, "unable to create your account", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          email,
		HashedPassword: hash,
	})

	if err != nil {
		msg := fmt.Sprintf("%s", err)
		log.Println(msg)
		respondWithError(w, http.StatusBadRequest, "Error Creating User", err)
		return

	}

	userResponse := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, 201, userResponse)

}
