package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pirateunclejack/go-practice/go-file-encryption/filecrypt"
	"golang.org/x/term"
)

func printHelp() {
	fmt.Println("file encryption")
	fmt.Println("Simple file encrypter for your day-to-day needs.")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("\tgo run . encrypt /path/to/your/file")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("\tencrypt\tEncrypts a file give a password")
	fmt.Println("\tdecrypt\tTries to decrypt a file using a password")
	fmt.Println("\thelp\tDisplays help text")
	fmt.Println("")
}


func validatePassword(password1 []byte, password2 []byte) bool {
	return bytes.Equal(password1, password2)
}

func getPassword() []byte {
	fmt.Print("Enter password: ")
	password, _ := term.ReadPassword(0)
	fmt.Print("\nConfirm Password: ")
	password2,_ := term.ReadPassword(0)
	if !validatePassword(password, password2) {
		fmt.Println("\nPasswords do not match. Please try again")
	}

	return password
}

func validateFile(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func encryptHandle() {
	if len(os.Args) < 3 {
		fmt.Println("missing the path to the file. For more infor, run go run . help")
		os.Exit(0)
	}

	file := os.Args[2]
	if !validateFile(file) {
		panic("File not found")
	}

	password := getPassword()
	fmt.Println("\nEncrypting...")
	filecrypt.Encrypt(file, password)
	fmt.Println("\n file successfully protected")
}

func decryptHandle() {
	if len(os.Args) < 3 {
		fmt.Println("missing the path to the file. For more infor, run go run . help")
		os.Exit(0)
	}

	file := os.Args[2]
	if !validateFile(file) {
		panic("File not found")
	}

	fmt.Print("Enter password: ")
	password, _ := term.ReadPassword(0)
	fmt.Println("\nDecrypting...")
	filecrypt.Decrypt(file, password)
	fmt.Println("\n file successfully decrypted")

}


func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	function := os.Args[1]

	switch function {
	case "help":
		printHelp()
	case "encrypt":
		encryptHandle()
	case "decrypt":
		decryptHandle()
	default:
		fmt.Println("Run encrypt to encrypt a file, and decrypt to decrypt a file.")
		os.Exit(1)
	}
}
