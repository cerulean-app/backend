package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: Unmark items that are done, repeat at a time before now and have an old updatedAt value.

type TodoData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
	Repeating   string `json:"repeating"`
	DueDate     string `json:"dueDate"`
}

func createTodoHandler(w http.ResponseWriter, r *http.Request, username string, token string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"error\":\"Invalid body sent!\"}", http.StatusBadRequest)
		return
	}
	var todo TodoData
	err = json.Unmarshal(body, &todo)
	if err != nil {
		http.Error(w, "{\"error\":\"Invalid body sent!\"}", http.StatusBadRequest)
		return
	} else if todo.Name == "" {
		http.Error(w, "{\"error\":\"Todo name is required!\"}", http.StatusBadRequest)
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
	if todo.DueDate != "" {
		todoDocument.DueDate, err = time.Parse("2006-01-02T15:04:05.999Z07:00", todo.DueDate)
		if err != nil {
			http.Error(w, "{\"error\":\"Invalid due date provided!\"}", http.StatusBadRequest)
			return
		}
	}
	result, err := database.Collection("users").UpdateOne(
		*mongoCtx, bson.M{"username": username}, bson.M{"$push": bson.M{"todos": todoDocument}},
	)
	if err != nil || result.MatchedCount != 1 {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
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
	result, err := database.Collection("users").UpdateOne(*mongoCtx, bson.M{"username": username}, bson.M{
		"$pull": bson.M{"todos": bson.M{"id": id}},
	})
	if err != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	} else if result.ModifiedCount == 0 {
		http.Error(w, "{\"error\":\"Todo not found!\"}", http.StatusNotFound)
		return
	}
	w.Write([]byte("{\"success\":true}")) // TODO: Inconsistent with documentation!
}

func patchTodoHandler(w http.ResponseWriter, r *http.Request, username string, id string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"error\":\"Invalid body sent!\"}", http.StatusBadRequest)
		return
	}
	var todo TodoData
	err = json.Unmarshal(body, &todo)
	if err != nil {
		http.Error(w, "{\"error\":\"Invalid body sent!\"}", http.StatusBadRequest)
		return
	}
	// TODO: Update updatedAt as well, and unmark done item that has to repeat and hasn't been updated yet.
	http.Error(w, "{\"error\":\"This endpoint is incomplete! Check back later.\"}", http.StatusServiceUnavailable)
	/*
		TODO: Complete implementation.
		result, err := database.Collection("users").UpdateOne(*mongoCtx, bson.M{"username": username}, bson.M{
			"$pull": bson.M{"todos": bson.M{"id": id}},
		})
		if err != nil {
			http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
			return
		} else if result.ModifiedCount == 0 {
			http.Error(w, "{\"error\":\"Todo not found!\"}", http.StatusNotFound)
			return
		}
		w.Write([]byte("{\"success\":true}")) // TODO: Inconsistent with documentation!
	*/
}

func getTodoHandler(w http.ResponseWriter, r *http.Request, username string, id string) {
	result := database.Collection("users").FindOne(*mongoCtx, bson.M{"username": username})
	if result.Err() != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	}
	var user UserDocument
	err := result.Decode(&user)
	if err != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	}
	for _, todo := range user.Todos {
		if todo.ID.Hex() == id {
			json.NewEncoder(w).Encode(todo)
			return
		}
	}
	http.Error(w, "{\"error\":\"Todo not found!\"}", http.StatusNotFound)
}

func getTodosHandler(w http.ResponseWriter, r *http.Request, username string, token string) {
	result := database.Collection("users").FindOne(*mongoCtx, bson.M{"username": username})
	if result.Err() != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	}
	var user UserDocument
	err := result.Decode(&user)
	if err != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user.Todos)
}
