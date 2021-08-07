package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO: Mark items that are done, repeat at a time before now and after updatedAt, as undone.

type TodoData struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Done        bool            `json:"done"`
	Repeating   string          `json:"repeating"`
	DueDate     json.RawMessage `json:"dueDate"`
}

func createTodoHandler(w http.ResponseWriter, r *http.Request, username string, token string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	}
	var todo TodoData
	err = json.Unmarshal(body, &todo)
	if err != nil {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	} else if todo.Name == "" {
		http.Error(w, `{"error":"Todo name is required!"}`, http.StatusBadRequest)
		return
	}
	nowTime := time.Now().UTC()
	todoDocument := TodoDocument{
		ID:          primitive.NewObjectIDFromTimestamp(nowTime),
		Name:        todo.Name,
		Description: todo.Description,
		Done:        todo.Done,
		Repeating:   todo.Repeating,
		CreatedAt:   nowTime,
		UpdatedAt:   nowTime,
	}
	if len(todo.DueDate) > 0 && string(todo.DueDate) != "null" {
		todoDocument.DueDate, err = time.Parse("2006-01-02T15:04:05.999Z07:00", string(todo.DueDate))
		if err != nil {
			http.Error(w, `{"error":"Invalid due date provided!"}`, http.StatusBadRequest)
			return
		}
	}
	result, err := database.Collection("users").UpdateOne(
		*mongoCtx, bson.M{"username": username}, bson.M{"$push": bson.M{"todos": todoDocument}},
	)
	if err != nil || result.MatchedCount != 1 {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(todoDocument)
}

func todoHandler(w http.ResponseWriter, r *http.Request, username string, token string) {
	pathSegments := strings.Split(r.URL.Path, "/")[2:]
	if len(pathSegments) != 1 {
		http.NotFound(w, r)
		return
	}
	id := pathSegments[0]
	if r.Method == "DELETE" {
		deleteTodoHandler(w, r, username, id)
	} else if r.Method == "PATCH" {
		patchTodoHandler(w, r, username, id)
	} else {
		getTodoHandler(w, r, username, id)
	}
}

func deleteTodoHandler(w http.ResponseWriter, r *http.Request, username string, id string) {
	result := database.Collection("users").FindOneAndUpdate(
		*mongoCtx,
		bson.M{"username": username, "todos": bson.M{"id": id}},
		bson.M{"$pull": bson.M{"todos": bson.M{"id": id}}},
	)
	if result.Err() == mongo.ErrNoDocuments {
		http.Error(w, `{"error":"Todo not found!"}`, http.StatusNotFound)
		return
	} else if result.Err() != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	var user UserDocument
	err := result.Decode(&user)
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	var foundTodo *TodoDocument
	for _, todo := range user.Todos {
		if todo.ID.Hex() == id {
			foundTodo = &todo
			break
		}
	}
	if foundTodo == nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(foundTodo)
}

func patchTodoHandler(w http.ResponseWriter, r *http.Request, username string, id string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	}
	var todo TodoData
	err = json.Unmarshal(body, &todo)
	if err != nil {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	}
	setOp := bson.M{
		"todos.$.done":      todo.Done,
		"todos.$.updatedAt": time.Now().UTC(),
	}
	if todo.Name != "" {
		setOp["todos.$.name"] = todo.Name
	}
	if todo.Description != "" {
		setOp["todos.$.description"] = todo.Description
	}
	update := bson.M{"$set": setOp}
	if len(todo.DueDate) > 0 && string(todo.DueDate) == "null" {
		update["$unset"] = bson.M{"dueDate": 1}
	} else if len(todo.DueDate) > 0 {
		setOp["todos.$.dueDate"], err = time.Parse("2006-01-02T15:04:05.999Z07:00", string(todo.DueDate))
		if err != nil {
			http.Error(w, `{"error":"Invalid due date provided!"}`, http.StatusBadRequest)
			return
		}
	}
	after := options.After
	result := database.Collection("users").FindOneAndUpdate(
		*mongoCtx, bson.M{"username": username, "todos.id": id}, update,
		&options.FindOneAndUpdateOptions{ReturnDocument: &after},
	)
	if result.Err() == mongo.ErrNoDocuments {
		http.Error(w, `{"error":"Todo not found!"}`, http.StatusNotFound)
		return
	} else if result.Err() != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusNotFound)
		return
	}
	var user UserDocument
	err = result.Decode(&user)
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	var foundTodo *TodoDocument
	for _, todo := range user.Todos {
		if todo.ID.Hex() == id {
			foundTodo = &todo
			break
		}
	}
	if foundTodo == nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(foundTodo)
}

func getTodoHandler(w http.ResponseWriter, r *http.Request, username string, id string) {
	result := database.Collection("users").FindOne(*mongoCtx, bson.M{"username": username})
	if result.Err() != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	var user UserDocument
	err := result.Decode(&user)
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	for _, todo := range user.Todos {
		if todo.ID.Hex() == id {
			json.NewEncoder(w).Encode(todo)
			return
		}
	}
	http.Error(w, `{"error":"Todo not found!"}`, http.StatusNotFound)
}

func getTodosHandler(w http.ResponseWriter, r *http.Request, username string, token string) {
	result := database.Collection("users").FindOne(*mongoCtx, bson.M{"username": username})
	if result.Err() != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	var user UserDocument
	err := result.Decode(&user)
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user.Todos)
}
