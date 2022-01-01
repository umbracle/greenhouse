package soltestlib

import _ "embed"

//go:generate go run gen/main.go

//go:embed testify.sol
var test string

//go:embed console.sol
var console string

func GetSolcTest() string {
	return test
}

func GetConsoleLib() string {
	return console
}
