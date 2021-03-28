package endpoints

import (
	"log"
	"net/http"
	"context"
	"firebase.google.com/go/v4/db"
	"encoding/json"

	"github.com/roblburris/pool-backend/database"
)

// UpdateProfile - Handler for updating a user profile
func UpdateProfile(ctx context.Context, ref *db.Ref) RequestHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Println("Error: expected Post request")

			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var user database.User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&user)
		if err != nil {
			text := "400 - couldn't parse JSON in request body"
			log.Printf("Error: %s Response: %s\n", err, text)

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Update the info in the database
		database.UpdateUser(ctx, ref, user)

		w.WriteHeader(http.StatusOK)
	}
}