package hqgossh

import "io"

// Command represents remote commands structure.
type Command struct { //nolint:govet // To be refactored.
	CMD    string
	ENV    map[string]string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}
