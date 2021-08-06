package main

import "net/http"

func todoHandler(w http.ResponseWriter, r *http.Request) {
	username := handleLoginCheck(w, r)
	if username == "" {
		return
	}
	// TODO: Read the body and create a todo.
}
