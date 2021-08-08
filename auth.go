package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/argon2"
)

func hashPassword(password string, salt string) string {
	return hex.EncodeToString(argon2.IDKey([]byte(password), []byte(salt), 18, 32*1024, 4, 32))
}

func generateToken() ([]byte, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, `{"error":"Allowed methods: POST"}`, http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	}
	var loginData LoginData
	err = json.Unmarshal(body, &loginData)
	if err != nil {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	}
	result := database.Collection("users").FindOne(mongoCtx, bson.M{"username": loginData.Username})
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		http.Error(w, `{"error":"Invalid username or password!"}`, http.StatusUnauthorized)
		return
	} else if result.Err() != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	// Hash and check the password.
	var user UserDocument
	err = result.Decode(&user)
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	} else if hashPassword(loginData.Password, user.Salt) != user.Password {
		http.Error(w, `{"error":"Invalid username or password!"}`, http.StatusUnauthorized)
		return
	} else if user.Verified != "" {
		http.Error(w, `{"error":"Account not verified!"}`, http.StatusUnauthorized)
		return
	}
	bytes, err := generateToken()
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	token := base64.StdEncoding.EncodeToString(bytes)
	_, err = database.Collection("tokens").InsertOne(mongoCtx, bson.M{
		"token":    token,
		"username": loginData.Username,
		"issuedOn": time.Now().UTC(),
	})
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
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

type RegisterData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, `{"error":"Allowed methods: POST"}`, http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	}
	var registerData RegisterData
	err = json.Unmarshal(body, &registerData)
	if err != nil {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	}
	result := database.Collection("users").FindOne(mongoCtx, bson.M{
		"$or": bson.A{bson.M{"username": registerData.Username}, bson.M{"email": registerData.Email}},
	})
	if result.Err() == nil {
		var user UserDocument
		err = result.Decode(&user)
		if err != nil {
			http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		} else if user.Email == registerData.Email {
			http.Error(w, `{"error":"A user with this email already exists!"}`, http.StatusConflict)
		} else {
			http.Error(w, `{"error":"A user with this username already exists!"}`, http.StatusConflict)
		}
		return
	} else if !errors.Is(result.Err(), mongo.ErrNoDocuments) {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	// Validate password and email.
	emailRegex, err := regexp.Compile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	} else if len(registerData.Password) < 8 {
		http.Error(w, `{"error":"Minimum password length: 8"}`, http.StatusBadRequest)
		return
	} else if !emailRegex.MatchString(registerData.Email) {
		http.Error(w, `{"error":"Invalid email provided!"}`, http.StatusBadRequest)
		return
	}
	saltBytes, err := generateToken()
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	salt := hex.EncodeToString(saltBytes)
	_, err = database.Collection("users").InsertOne(mongoCtx, bson.M{
		"username":   registerData.Username,
		"password":   hashPassword(registerData.Password, salt),
		"email":      registerData.Email,
		"salt":       salt,
		"verified":   "",
		"lastEdited": time.Now().UTC(),
		"todos":      bson.A{},
	})
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	// Log the user in for now until email verification is added.
	bytes, err := generateToken()
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	token := base64.StdEncoding.EncodeToString(bytes)
	_, err = database.Collection("tokens").InsertOne(mongoCtx, bson.M{
		"token":    token,
		"username": registerData.Username,
		"issuedOn": time.Now().UTC(),
	})
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
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

func isLoggedIn(token string) (string, error) {
	result := database.Collection("tokens").FindOne(mongoCtx, bson.M{"token": token})
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return "", nil
	} else if result.Err() != nil {
		return "", result.Err()
	}
	var document TokenDocument
	err := result.Decode(&document)
	if err != nil {
		return "", err
	}
	// TODO: Idle timeout?
	if document.IssuedOn.UTC().Add(time.Hour * 24 * 180).Before(time.Now().UTC()) {
		_, _ = database.Collection("tokens").DeleteOne(mongoCtx, bson.M{"token": token})
		return "", nil
	}
	return document.Username, nil
}

func handleLoginCheck(
	handler func(w http.ResponseWriter, r *http.Request, username string, token string),
	methods []string,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		allowedMethod := false
		for _, method := range methods {
			if r.Method == method {
				allowedMethod = true
			}
		}
		if !allowedMethod {
			http.Error(w, `{"error":"Allowed methods: `+strings.Join(methods, ", ")+`"}`, http.StatusMethodNotAllowed)
			return
		}
		cookie, err := r.Cookie("cerulean_token")
		var token string
		if errors.Is(err, http.ErrNoCookie) {
			token = r.Header.Get("Authorization")
		} else {
			token = cookie.Value
		}
		if token == "" {
			http.Error(w, `{"error":"No access token provided!"}`, http.StatusUnauthorized)
			return
		}
		username, err := isLoggedIn(token)
		if err != nil {
			http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
			return
		} else if username == "" {
			http.Error(w, `{"error":"Invalid access token provided!"}`, http.StatusUnauthorized)
			return
		}
		handler(w, r, username, token)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, `{"error":"Allowed methods: POST"}`, http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("cerulean_token")
	var token string
	if errors.Is(err, http.ErrNoCookie) {
		token = r.Header.Get("Authorization")
	} else {
		token = cookie.Value
	}
	if token == "" {
		http.Error(w, `{"error":"No access token provided!"}`, http.StatusUnauthorized)
		return
	}
	result, err := database.Collection("tokens").DeleteOne(mongoCtx, bson.M{"token": token})
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	} else if result.DeletedCount == 0 {
		http.Error(w, `{"error":"Invalid access token provided!"}`, http.StatusUnauthorized)
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
	w.Write([]byte(`{"success":true}`))
}

type ChangePasswordData struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func changePasswordHandler(w http.ResponseWriter, r *http.Request, username string, token string) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"Allowed methods: POST"}`, http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	}
	var passwordData ChangePasswordData
	err = json.Unmarshal(body, &passwordData)
	if err != nil || passwordData.NewPassword == "" || passwordData.CurrentPassword == "" {
		http.Error(w, `{"error":"Invalid body sent!"}`, http.StatusBadRequest)
		return
	}
	result := database.Collection("users").FindOne(mongoCtx, bson.M{"username": username})
	if result.Err() != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	var user UserDocument
	err = result.Decode(&user)
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	} else if hashPassword(passwordData.CurrentPassword, user.Salt) != user.Password {
		http.Error(w, `{"error":"Invalid password!"}`, http.StatusUnauthorized)
		return
	}
	saltBytes, err := generateToken()
	if err != nil {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	salt := hex.EncodeToString(saltBytes)
	updateResult, err := database.Collection("users").UpdateOne(
		mongoCtx, bson.M{"username": username}, bson.M{"$set": bson.M{
			"password": hashPassword(passwordData.NewPassword, salt), "salt": salt,
		}},
	)
	if err != nil || updateResult.ModifiedCount != 1 {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	w.Write([]byte(`{"success":true}`))
}

func deleteAccountHandler(w http.ResponseWriter, r *http.Request, username string, token string) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"Allowed methods: POST"}`, http.StatusMethodNotAllowed)
		return
	}
	result, err := database.Collection("users").DeleteOne(mongoCtx, bson.M{"username": username})
	if err != nil || result.DeletedCount != 1 {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
		return
	}
	result, err = database.Collection("tokens").DeleteMany(mongoCtx, bson.M{"username": username})
	if err != nil || result.DeletedCount <= 0 {
		http.Error(w, `{"error":"Internal Server Error!"}`, http.StatusInternalServerError)
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
	w.Write([]byte(`{"success":true}`))
}
