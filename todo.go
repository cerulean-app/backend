package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TodoData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
	Repeating   string `json:"repeating"`
	DueDate     string `json:"dueDate"`
}

func todoHandler(w http.ResponseWriter, r *http.Request, username string) {
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
	result, err := database.Collection("Users").UpdateOne(*mongoCtx, bson.M{"username": username}, bson.M{
		"$push": bson.M{"todos": todoDocument},
	})
	if err != nil || result.MatchedCount != 1 {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(todoDocument)
}
