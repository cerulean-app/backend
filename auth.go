package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/argon2"
)

func hashPassword(password string, salt string) string {
	return string(argon2.IDKey([]byte(password), []byte(salt), 1, 8*1024, 4, 32))
}

type JwtClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func createJwt(claims JwtClaims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.JwtSecret)
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
	var user UsersCollectionDocument
	err = result.Decode(&user)
	if err != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	} else if hashPassword(loginData.Password, user.Salt) != user.Password {
		http.Error(w, "{\"error\":\"Invalid username or password!\"}", http.StatusUnauthorized)
		return
	}
	jwt, err := createJwt(JwtClaims{
		Username:       user.Username,
		StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix()},
	})
	if err != nil {
		http.Error(w, "{\"error\":\"Internal Server Error!\"}", http.StatusInternalServerError)
		return
	}
	// TODO: Add Secure to cookie.
	w.Header().Add("Set-Cookie", "cerulean-token="+jwt+"; HttpOnly; SameSite=Lax; Max-Age=31536000")
	json.NewEncoder(w).Encode(map[string]string{"token": jwt})
}
