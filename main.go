package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
	Port     int    `json:"port"`
	MongoUri string `json:"mongoUri"`
}

func main() {
	// Read config.
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		panic(err)
	}
	// Connect to MongoDB.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoUri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	// Ping the primary MongoDB server to ensure the connection is good.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to MongoDB.")

	// Create CORS handler wrapper.
	corsHandlerWrapper := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Accept"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "PATCH", "DELETE"}),
		handlers.AllowedOrigins([]string{"*"}),
	)
	http.Handle("/", corsHandlerWrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"Hello, World!"}`))
	})))

	// Start listening on specified port.
	fmt.Println("Listening on port", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
	// if !config.HTTPS.Enabled { err = http.ListenAndServe(config.Port, handler) }
	// else { err = http.ListenAndServeTLS(port, config.HTTPS.Cert, config.HTTPS.Key, handler) }
}
