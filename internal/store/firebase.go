package store

import (
	"bytes"
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"io"
	"log"
	"net/http"
	"os"
)

var App *firebase.App
var Ctx = context.Background()

//go:embed simple-ssh-manager-firebase-adminsdk-y7ei5-ac0913e54f.json
var firebaseConfig embed.FS

// InitFirebase initializes the Firebase app and assigns it to App
func InitFirebase() error {
	configFile, err := firebaseConfig.ReadFile("simple-ssh-manager-firebase-adminsdk-y7ei5-ac0913e54f.json")
	if err != nil {
		return fmt.Errorf("error reading embedded config file: %v", err)
	}

	opt := option.WithCredentialsJSON(configFile)
	App, err = firebase.NewApp(Ctx, nil, opt)
	if err != nil {
		logrus.Fatalf("init firebase failed: %v", err)
	}

	return nil
}

// RegisterUser registers a new user with email and password
func RegisterUser(email, password string) (*auth.UserRecord, error) {
	client, err := App.Auth(Ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %v", err)
	}
	params := (&auth.UserToCreate{}).
		Email(email).
		Password(password)
	user, err := client.CreateUser(Ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}
	return user, nil
}

// LoginUser authenticates a user with email and password
func LoginUser(email, password string) (map[string]interface{}, error) {
	_, err := App.Auth(Ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %v", err)
	}
	// Firebase Admin SDK does not support direct login; use custom tokens or an external client for login.
	token, err := authenticateWithFirebase(email, password)
	if err != nil {
		return nil, fmt.Errorf("⚠️ %v", err)
	}
	return token, nil
}

func authenticateWithFirebase(email, password string) (map[string]interface{}, error) {
	ApiKey := os.Getenv("API_KEY")
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", ApiKey)
	payload := map[string]string{
		"email":             email,
		"password":          password,
		"returnSecureToken": "true",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong Credentials: %d", resp.StatusCode)
	}

	var response map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	idToken, ok := response["idToken"].(string)
	if !ok {
		return nil, fmt.Errorf("error extracting ID token from response")
	}
	return parseToken(idToken), nil
}

func parseToken(tokenString string) map[string]interface{} {
	// Parse the token without verification
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Fatalf("Error parsing token: %v", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		result := make(map[string]interface{})
		if email, ok := claims["email"]; ok {
			result["email"] = email
		}
		if userID, ok := claims["user_id"]; ok {
			result["user_id"] = userID
		}
		return result
	} else {
		log.Fatalf("Error parsing claims")
		return nil
	}
}
