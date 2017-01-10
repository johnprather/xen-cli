package main

import "log"

var secure Secure

// Secure is an interface for functions to access various OS's keychains
// secure_darwin.go provides this interface for OSX keychain
// other files need to be made for kde wallet, gnome keyring, etc
type Secure interface {
	GetDefaultPassword() (pass string, err error)
	GetPassword(name string) (pass string, err error)
	SetDefaultPassword(pass string) (err error)
	SetPassword(name string, pass string) (err error)
}

// this gets called by main() to ensure we have a secure interface
func checkSecure() {
	if secure == nil {
		log.Fatalln("No valid security interface initialized.")
	}
}
