package main

type configClass struct {
	baseDir    string
	socketPath string
}

var config = &configClass{}
