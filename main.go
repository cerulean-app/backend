package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var config Config
var mongodb *mongo.Client
var database *mongo.Database
var mongoCtx *context.Context

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
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		panic(err)
	}
	// Connect to MongoDB.
	mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongodb, err = mongo.Connect(mongoCtx, options.Client().ApplyURI(config.MongoUri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = mongodb.Disconnect(mongoCtx); err != nil {
			panic(err)
		}
	}()

	// Define MongoDB schemas.
	database = mongodb.Database("cerulean")
	database.CreateCollection(mongoCtx, "users", &options.CreateCollectionOptions{
		Validator: bson.M{"$jsonSchema": UsersCollectionSchema},
	})
	database.CreateCollection(mongoCtx, "tokens", &options.CreateCollectionOptions{
		Validator: bson.M{"$jsonSchema": TokensCollectionSchema},
	})
	fmt.Println("Successfully connected to MongoDB.")

	// Create CORS handler wrapper.
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Accept"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "PATCH", "DELETE"}),
		handlers.AllowedOrigins([]string{"*"}),
	)
	// TODO: Switch to raw string literals in http.Error and w.Write calls.
	// Authentication endpoints.
	http.Handle("/login", cors(http.HandlerFunc(loginHandler)))
	http.Handle("/logout", cors(http.HandlerFunc(logoutHandler)))
	http.Handle("/changepassword", cors(http.HandlerFunc(handleLoginCheck(changePasswordHandler, []string{"POST"}))))
	// Data endpoints.
	http.Handle("/todo", cors(http.HandlerFunc(handleLoginCheck(createTodoHandler, []string{"POST"}))))
	http.Handle("/todos", cors(http.HandlerFunc(handleLoginCheck(getTodosHandler, []string{"GET"}))))
	http.Handle("/todo/", cors(http.HandlerFunc(handleLoginCheck(todoHandler, []string{"DELETE", "PATCH", "GET"}))))

	// Start listening on specified port.
	fmt.Println("Listening on port", config.Port)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
	// if config.HTTPS.Enabled { http.ListenAndServeTLS(config.HTTPS.Cert, config.HTTPS.Key) }
}
