package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/argon2"
)

func hashPassword(password string, salt string) string {
	return string(argon2.IDKey([]byte(password), []byte(salt), 1, 8*1024, 4, 32))
}

func generateToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(token), nil
}

type LoginData struct {
	Username string
	Password string
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "{\"error\":\"Invalid body sent!\"}", http.StatusBadRequest)
		return
	}
	var loginData LoginData
	err = json.Unmarshal(body, &loginData)
	if err != nil {
		http.Error(w, "{\"error\":\"Invalid body sent!\"}", http.StatusBadRequest)
		return
	}
	result := database.Collection("users").FindOne(*mongoCtx, bson.M{"username": loginData.Username})
	if result.Err() == mongo.ErrNoDocuments {
		http.Error(w, "{\"error\":\"Invalid username or password!\"}", http.StatusUnauthorized)
		return
	} else if result.Err() != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	}
	// Hash and check the password.
	var user UserDocument
	err = result.Decode(&user)
	if err != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	} else if hashPassword(loginData.Password, user.Salt) != user.Password {
		http.Error(w, "{\"error\":\"Invalid username or password!\"}", http.StatusUnauthorized)
		return
	}
	token, err := generateToken()
	if err != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	}
	_, err = database.Collection("tokens").InsertOne(*mongoCtx, bson.M{
		"token":    token,
		"username": loginData.Username,
		"issuedOn": time.Now().UTC(),
	})
	if err != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("cookie") != "false" {
		// TODO: Add Secure to cookie.
		r.AddCookie(&http.Cookie{
			Name:     "cerulean_token",
			Value:    token,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   31536000,
		})
	}
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func isLoggedIn(token string) (string, error) { // TODO: IssuedOn expiry.
	result := database.Collection("tokens").FindOne(*mongoCtx, bson.M{"token": token})
	if result.Err() == mongo.ErrNoDocuments {
		return "", nil
	} else if result.Err() != nil {
		return "", result.Err()
	}
	var document TokenDocument
	result.Decode(&document)
	return document.Username, nil
}

// TODO: Use a higher-order function/
func handleLoginCheck(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("cerulean_token")
	var token string
	if err == http.ErrNoCookie {
		token = r.Header.Get("Authorization")
	} else {
		token = cookie.Value
	}
	if token == "" {
		http.Error(w, "{\"error\": \"No access token provided!\"}", http.StatusUnauthorized)
		return ""
	}
	username, err := isLoggedIn(token)
	if err != nil {
		http.Error(w, "{\"error\": \"Internal Server Error!\"}", http.StatusInternalServerError)
		return ""
	} else if username == "" {
		http.Error(w, "{\"error\": \"Invalid access token provided!\"}", http.StatusUnauthorized)
		return ""
	}
	return username
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cookie, err := r.Cookie("cerulean_token")
	var token string
	if err == http.ErrNoCookie {
		token = r.Header.Get("Authorization")
	} else {
		token = cookie.Value
	}
	if token == "" {
		http.Error(w, "{\"error\": \"No access token provided!\"}", http.StatusUnauthorized)
		return
	}
	result, err := database.Collection("tokens").DeleteOne(*mongoCtx, bson.M{"token": token})
	if err != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	} else if result.DeletedCount == 0 {
		http.Error(w, "{\"error\":\"Invalid access token provided!\"}", http.StatusUnauthorized)
		return
	}
	if _, err = r.Cookie("cerulean_token"); err != http.ErrNoCookie {
		r.AddCookie(&http.Cookie{
			Name:     "cerulean_token",
			Value:    "",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   1,
		})
	}
	w.Write([]byte("{\"success\":true}"))
}
