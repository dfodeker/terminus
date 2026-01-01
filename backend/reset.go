package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}

	//cfg.fileServerHits.Store(0)
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the database: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and database reset to initial state."))
}

// func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {

// 	if cfg.platform != "dev" {
// 		respondWithError(w, 403, "Forbidden")
// 	}
// cfg.fileServerHits.Store(0)
// 	err := cfg.db.DeleteAllUsers(r.Context())
// 	if err != nil {
// 		msg := fmt.Sprintf("%s", err)
// 		respondWithError(w, 403, "Error Occured while resetting: "+msg)
// 	}

// 	respondWithJson(w, 200, "")
// }
