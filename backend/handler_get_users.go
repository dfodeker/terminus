package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerGetUsers(w http.ResponseWriter, r *http.Request) {
	//typically you'd want only authenticated users to be able to veiw , we also dont want to be able to view all users so
	//so when we add in store restrictions we'll limit that too
	response := []User{}
	users, err := cfg.db.GetAllUsers(r.Context())
	if err != nil {
		log.Println(err)
		respondWithError(w, http.StatusServiceUnavailable, "Unable to retrieve users ", err)
		return
	}
	//respond with results count

	for _, u := range users {
		response = append(response, User{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			Email:     u.Email,
		})

	}

	respondWithJSON(w, http.StatusOK, response)

}
