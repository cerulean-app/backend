package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var UsersCollectionSchema = bson.M{
	"required": []string{"username", "email", "password"},
	"properties": bson.M{
		"username": bson.M{
			"bsonType":  "string",
			"minLength": 4,
		},
		"password": bson.M{
			"bsonType":  "string",
			"minLength": 16,
		},
		"salt": bson.M{
			"bsonType":  "string",
			"minLength": 16,
		},
		"email": bson.M{
			"bsonType":  "string",
			"minLength": 4,
			"pattern":   "^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$",
		},
	},
}

type UsersCollectionDocument struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Salt     string `json:"salt" bson:"salt"`
	Email    string `json:"email" bson:"email"`
}

var TokensCollectionSchema = bson.M{
	"required": []string{"username", "accessToken", "issuedOn"},
	"properties": bson.M{
		"token":    bson.M{"bsonType": "string", "minLength": 42},
		"username": bson.M{"bsonType": "string", "minLength": 4},
		"issuedOn": bson.M{"bsonType": "date"},
	},
}

type TokensCollectionDocument struct {
	Username string    `json:"username" bson:"username"`
	IssuedOn time.Time `json:"issuedOn" bson:"issuedOn"`
	Token    string    `json:"token" bson:"token"`
}
