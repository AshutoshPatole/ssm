package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

func AskPassword() (string, error) {
	fmt.Print("Password: ")
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Print("\n")
	return string(password), nil
}
