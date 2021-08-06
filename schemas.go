package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var UsersCollectionSchema = bson.M{
	"required": []string{"username", "email", "password", "salt", "todos", "lastEdited"},
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
		"lastEdited": bson.M{"bsonType": "date"},
		"todos": bson.M{
			"bsonType": "array",
			"items": bson.M{
				"bsonType": "object",
				"required": []string{"id", "name", "description", "done", "repeating", "createdAt", "updatedAt"},
				"properties": bson.M{
					"id":          bson.M{"bsonType": "objectId"},
					"name":        bson.M{"bsonType": "string", "minLength": 1},
					"description": bson.M{"bsonType": "string"},
					"done":        bson.M{"bsonType": "boolean"},
					"createdAt":   bson.M{"bsonType": "date"},
					"updatedAt":   bson.M{"bsonType": "date"},
					"repeating":   bson.M{"bsonType": "string", "enum": []string{"daily", "weekly", "monthly", "yearly"}},
					"dueDate":     bson.M{"bsonType": "date"},
				},
			},
		},
	},
}

type UserDocument struct {
	Username   string         `json:"username" bson:"username"`
	Password   string         `json:"password" bson:"password"`
	Salt       string         `json:"salt" bson:"salt"`
	Email      string         `json:"email" bson:"email"`
	LastEdited time.Time      `json:"lastEdited" bson:"lastEdited"`
	Todos      []TodoDocument `json:"todos" bson:"todos"`
}

type TodoDocument struct {
	ID          primitive.ObjectID `json:"id" bson:"id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Done        bool               `json:"done" bson:"done"`
	Repeating   string             `json:"repeating" bson:"repeating"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	DueDate     time.Time          `json:"dueDate" bson:"dueDate"`
}

var TokensCollectionSchema = bson.M{
	"required": []string{"username", "accessToken", "issuedOn"},
	"properties": bson.M{
		"token":    bson.M{"bsonType": "string", "minLength": 42},
		"username": bson.M{"bsonType": "string", "minLength": 4},
		"issuedOn": bson.M{"bsonType": "date"},
	},
}

type TokenDocument struct {
	Username string    `json:"username" bson:"username"`
	IssuedOn time.Time `json:"issuedOn" bson:"issuedOn"`
	Token    string    `json:"token" bson:"token"`
}
