package ssh

import "io"

// Command represents remote commands structure.
type Command struct {
	CMD    string
	ENV    map[string]string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}
