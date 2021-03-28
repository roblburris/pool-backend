package database

import (
	"log"
	"fmt"
	"context"
	"firebase.google.com/go/v4/db"
)


// GetUsers - gets all users from db 
// Returns a map[int64]User that maps a user id to a specific user (see User struct)
func GetUsers(ctx context.Context, ref *db.Ref) map[int64]User {
	var data map[int64]User
	if err := ref.Get(ctx, &data); err != nil {
		log.Println("Error reading from database:", err)
		return nil
	}

	return data
}

// AddUser - adds user to DB
func AddUser() {
	return
}

// UpdatePair - updates the pairings given a map of pairings
func UpdatePair(ctx context.Context, ref *db.Ref, pairings map[int64]int64) {
	for k, v := range pairings {
		if err := ref.Update(ctx, map[string]interface{}{
			fmt.Sprintf("%d/partner", k): v,
			fmt.Sprintf("%d/partner", v): k,
		}); err != nil {
			log.Println("Error updating info:", err)
		}
	}
}

// UpdateUser - updates a user with the new info passed in
func UpdateUser(ctx context.Context, ref *db.Ref, newInfo User) {
	if err := ref.Update(ctx, map[string]interface{} {
		fmt.Sprintf("%d", newInfo.ID): newInfo,
	}); err != nil {
		log.Println("Error updating user ", newInfo.ID, err)	
	}
}
