package main

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"time"

	"github.com/dfodeker/terminus/internal/auth"
	"github.com/dfodeker/terminus/internal/database"
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
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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
		return
	}

	email := params.Email
	_, err = mail.ParseAddress(params.Email)
	if err != nil {
		respondWithError(w, 400, "please provide a valid email", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	//compare password
	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email of password", err)
		return

	}
	refreshTime := time.Now().Add(60 * 24 * time.Hour)

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal server error", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.signingKey, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal server error", err)
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: refreshTime,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to generate token", err)
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}
