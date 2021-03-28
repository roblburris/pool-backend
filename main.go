package main

import (
	"log"
	"context"
	"io/ioutil"
  
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"googlemaps.github.io/maps"

	"github.com/roblburris/pool-backend/database"
  )
  
func main() {
	ctx := context.Background()
	conf := &firebase.Config{
			DatabaseURL: "https://carpool-app-e4a60-default-rtdb.firebaseio.com/",
	}
	opt := option.WithCredentialsFile("./auth-keys/carpool-app-e4a60-firebase-adminsdk-8jiru-d26484ceb9.json")

	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalln("error initializing app:", err)
	}

	client, err := app.Database(ctx)
	if err != nil {
			log.Fatalln("Error initializing database client:", err)
	}

	
	ref := client.NewRef("users/")
	database.GetUsers(ctx, ref)

	// Create a Maps API Client
	mapsAPIKey, err := ioutil.ReadFile("./auth-keys/maps-api.txt")
	if err != nil {
		log.Fatal(err)
	}
	c, err := maps.NewClient(maps.WithAPIKey(string(mapsAPIKey)))

	log.Println(c)
	database.UpdatePair(ctx, ref, 436587342, 1216595250)
}
