package store

import (
	"context"
	"firebase.google.com/go"
	"google.golang.org/api/option"
)

var App *firebase.App
var Ctx = context.Background()

// InitFirebase initializes the Firebase app and assigns it to App
func InitFirebase() error {
	if App != nil {
		return nil // Firebase app is already initialized
	}
	opt := option.WithCredentialsFile("/home/ashu/ssm-v2/simple-ssh-manager-firebase-adminsdk-y7ei5-ac0913e54f.json")
	var err error
	App, err = firebase.NewApp(Ctx, nil, opt)
	if err != nil {
		return err
	}
	return nil
}
