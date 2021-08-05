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
	Port      int    `json:"port"`
	MongoUri  string `json:"mongoUri"`
	JwtSecret string `json:"jwtSecret"`
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
	http.Handle("/login", corsHandlerWrapper(http.HandlerFunc(loginHandler)))

	// Start listening on specified port.
	fmt.Println("Listening on port", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
	// if !config.HTTPS.Enabled { err = http.ListenAndServe(config.Port, handler) }
	// else { err = http.ListenAndServeTLS(port, config.HTTPS.Cert, config.HTTPS.Key, handler) }
}
