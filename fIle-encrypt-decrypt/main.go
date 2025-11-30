package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/MarufHossain14/golang-projects/file-encrypt-decrypt/cryption"

	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}
	function := os.Args[1]

	switch function {
	case "help":
		printHelp()
	case "encrypt":
		encrypthandle()
	case "decrypt":
		decrypthandle()
	default:
		fmt.Println("Unknown command:")
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("file encryption")
	fmt.Println("Simple file encryption and decryption using AES-GCM")
	fmt.Println("Usage:")
	fmt.Println("")
	fmt.Println("\tgo run . encrypt /path/to/source/file")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("")
	fmt.Println("\t encrypt \t Encrypt the specified file")
	fmt.Println("\t decrypt \t Decrypt the specified file")
	fmt.Println("\t help \t\t Show this help message")
	fmt.Println("")
}

func encrypthandle() {
	if len(os.Args) < 3 {
		println("Please provide the path to the file to encrypt, run go run . help")
		os.Exit(0)
	}
	file := os.Args[2]
	if !validateFile(file) {
		panic("File does not exist")
	}
	password := getPassword()
	fmt.Println("\nEncrypting file:")
	cryption.Encrypt(file, password)
	fmt.Println("File encrypted successfully")
}

func decrypthandle() {
	if len(os.Args) < 3 {
		println("Please provide the path to the file to decrypt, run go run . help")
		os.Exit(1)
	}
	file := os.Args[2]
	if !validateFile(file) {
		panic("File does not exist")
	}
	fmt.Print("Enter password: ")
	password, _ := term.ReadPassword(0)
	fmt.Println("\nDecrypting file:")
	cryption.Decrypt(file, password)
	fmt.Println("File decrypted successfully")
}

func getPassword() []byte {
	fmt.Print("Enter password: ")
	password, _ := term.ReadPassword(0)
	fmt.Println("\nConfirm password: ")
	confirmPassword, _ := term.ReadPassword(0)
	if string(password) != string(confirmPassword) {
		fmt.Print("Passwords do not match")
		return getPassword()

	}
	return password

}

func validateFile(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func validatePassworcd(password []byte, confirmPassword []byte) bool {
	if !bytes.Equal(password, confirmPassword) {
		return false
	}
	return true
}
