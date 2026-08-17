package security

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const appName = "SSM"

func StoreCredentials(key, value string) error {
	return keyring.Set(appName, key, value)
}

func RetreiveCredentials(key string) (string, error) {
	password, err := keyring.Get(appName, key)
	if err != nil {
		return "", fmt.Errorf("error fetching credentials %v", err)
	}
	return password, nil
}
