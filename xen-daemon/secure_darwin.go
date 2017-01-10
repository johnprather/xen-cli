// This is a set of secure
// +build darwin

package main

import "github.com/johnprather/go-simple-keychain/simpleKeychain"

// SecureDarwin is the security suite for OSX
type SecureDarwin struct {
	group   string
	account string
	prefix  string
}

func init() {
	secure = NewSecureDarwin()
}

// NewSecureDarwin returns an instantiated SecureDarwin object
func NewSecureDarwin() *SecureDarwin {
	secure := &SecureDarwin{}
	secure.group = "group.com.github.johnprather.xen-cli"
	secure.account = "root"
	secure.prefix = "com.github.johnprather.xen-cli."
	return secure
}

// GetDefaultPassword returns the default pw for new xapi servers
func (s *SecureDarwin) GetDefaultPassword() (pass string, err error) {
	return s.GetPassword("default")
}

// GetPassword returns the pw for specified xapi server
func (s *SecureDarwin) GetPassword(name string) (pass string, err error) {
	pass, err = simpleKeychain.Load(s.group, s.prefix+name, s.account)
	return
}

// SetDefaultPassword sets the default pass for xapi servers
func (s *SecureDarwin) SetDefaultPassword(pass string) (err error) {
	return s.SetPassword("default", pass)
}

// SetPassword sets the password for a specified xapi server
func (s *SecureDarwin) SetPassword(name string, pass string) (err error) {
	return simpleKeychain.Save(s.group, s.prefix+name, s.account, pass)
}

// DelPassword removes the password for specified xapi server
func (s *SecureDarwin) DelPassword(name string) (err error) {
	return simpleKeychain.Delete(s.group, s.prefix+name, s.account)
}
