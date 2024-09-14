package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

var App *firebase.App
var Ctx = context.Background()

// InitFirebase initializes the Firebase app and assigns it to App
func InitFirebase() error {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("error unmarshaling config: %v", err)
	}

	if config.FirebaseConfig == "" {
		return fmt.Errorf("Firebase configuration is missing in the config file")
	}

	configFile, err := os.ReadFile(config.FirebaseConfig)
	if err != nil {
		return fmt.Errorf("cannot find or read Firebase config file at %s: %v", config.FirebaseConfig, err)
	}

	opt := option.WithCredentialsJSON(configFile)
	App, err = firebase.NewApp(Ctx, nil, opt)
	if err != nil {
		logrus.Fatalf("init firebase failed: %v", err)
	}

	return nil
}

var (
	firebaseInitialized bool
	firebaseInitMutex   sync.Mutex
)

// InitFirebaseOnce initializes Firebase only once
func InitFirebaseOnce() error {
	firebaseInitMutex.Lock()
	defer firebaseInitMutex.Unlock()

	if !firebaseInitialized {
		if err := InitFirebase(); err != nil {
			return err
		}
		firebaseInitialized = true
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
	ApiKey := viper.GetString("firebaseApiKey")
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errorDetails, ok := errorResponse["error"].(map[string]interface{}); ok {
				return nil, fmt.Errorf("HTTP Error %d: %s", resp.StatusCode, errorDetails["message"])
			}
		}
		return nil, fmt.Errorf("HTTP Error %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
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

// ResetPassword sends a password reset email to the specified user
func ResetPassword(email string) error {
	client, err := App.Auth(Ctx)
	if err != nil {
		return fmt.Errorf("error getting Auth client: %v", err)
	}

	// Check if the user exists
	_, err = client.GetUserByEmail(Ctx, email)
	if err != nil {
		return fmt.Errorf("no user found with the email address: %s", email)
	}

	ApiKey := viper.GetString("firebaseApiKey")
	url := fmt.Sprintf("https://www.googleapis.com/identitytoolkit/v3/relyingparty/getOobConfirmationCode?key=%s", ApiKey)
	payload := map[string]string{
		"requestType": "PASSWORD_RESET",
		"email":       email,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error sending reset request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errorDetails, ok := errorResponse["error"].(map[string]interface{}); ok {
				return fmt.Errorf("HTTP Error %d: %s", resp.StatusCode, errorDetails["message"])
			}
		}
		return fmt.Errorf("HTTP Error %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Please check your %s inbox for password reset link.\n", email)
	return nil
}
