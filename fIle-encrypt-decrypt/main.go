package main

import (
	"fmt"
	"os"
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
	fmt.Println("encrypt handler not implemented")
}

func decrypthandle() {
	fmt.Println("decrypt handler not implemented")
}

func getPassword() {

}

func validatePassworcd() {

}

func validateFile() {

}
