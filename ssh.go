package hqgossh

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/hueristiq/hqgossh/authentication"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Client wraps SSH and SFTP clients.
type Client struct {
	SSH  *ssh.Client
	SFTP *sftp.Client
}

// Options represents the options required to establish a SSH/SFTP connection.
type Options struct { //nolint:govet // To be refactored.
	Host            string
	Port            int
	User            string
	Authentication  authentication.Authentication
	HostKeyCallback ssh.HostKeyCallback
}

// New creates a new Client with provided Options.
// It establishes both SSH and SFTP clients.
func New(options *Options) (client *Client, err error) {
	client = &Client{}

	server := net.JoinHostPort(
		options.Host,
		fmt.Sprint(options.Port),
	)
	config := &ssh.ClientConfig{
		User:            options.User,
		Auth:            options.Authentication,
		HostKeyCallback: options.HostKeyCallback,
	}

	if err := establishSSHConnection(client, server, config); err != nil {
		return nil, fmt.Errorf("failed establishing SSH connection: %s", err)
	}

	if err := createSFTPClient(client); err != nil {
		return nil, fmt.Errorf("failed creating SFTP client: %s", err)
	}

	return
}

// Run runs remote commands over SSH.
func (client *Client) Run(command *Command) (err error) {
	var (
		session *ssh.Session
	)

	session, err = client.SSH.NewSession()
	if err != nil {
		return
	}

	defer session.Close()

	if err = handleIOStreams(session, command); err != nil {
		return
	}

	if err = setEnvVariables(session, command.ENV); err != nil {
		return
	}

	if err = setPty(session); err != nil {
		return
	}

	if err = session.Run(command.CMD); err != nil {
		return
	}

	return
}

// Shell opens an interactive shell over SSH.
func (client *Client) Shell() (err error) {
	var (
		session *ssh.Session
	)

	session, err = client.SSH.NewSession()
	if err != nil {
		return
	}

	defer session.Close()

	session.Stdin, session.Stdout, session.Stderr = os.Stdin, os.Stdout, os.Stderr

	sTerm := os.Getenv("TERM")
	if sTerm == "" {
		sTerm = "xterm-256color"
	}

	termmodes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		ssh.OPOST:         1,     // Enable output processing.
	}

	fd := int(os.Stdin.Fd())

	state, err := term.MakeRaw(fd)
	if err != nil {
		return
	}

	defer func() {
		err = term.Restore(fd, state)
		if err != nil {
			return
		}
	}()

	width, height, err := term.GetSize(fd)
	if err != nil {
		return
	}

	if err = session.RequestPty(sTerm, height, width, termmodes); err != nil {
		return
	}

	if err = session.Shell(); err != nil {
		return
	}

	if err = session.Wait(); err != nil {
		return
	}

	return
}

// Close closes the SFTP and SSH clients.
func (client *Client) Close() (err error) {
	if client == nil {
		err = fmt.Errorf("failed closing: client is nil")

		return
	}

	if client.SFTP != nil {
		if err = client.SFTP.Close(); err != nil {
			err = fmt.Errorf("failed closing SFTP client: %s", err)

			return
		}
	}

	if client.SSH != nil {
		if err = client.SSH.Close(); err != nil {
			err = fmt.Errorf("failed closing SSH connection: %s", err)

			return
		}
	}

	return
}

// establishSSHConnection establishes a SSH connection.
func establishSSHConnection(client *Client, server string, config *ssh.ClientConfig) (err error) {
	maxRetries := 3
	retryDelay := 10 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		client.SSH, err = ssh.Dial("tcp", server, config)
		if err == nil {
			return
		}

		time.Sleep(retryDelay)
	}

	return
}

// createSFTPClient creates a SFTP client.
func createSFTPClient(client *Client) (err error) {
	maxRetries := 3
	retryDelay := 10 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		client.SFTP, err = sftp.NewClient(client.SSH)
		if err == nil {
			return
		}

		time.Sleep(retryDelay)
	}

	return
}

// handleIOStreams handles I/O streams for the command.
func handleIOStreams(session *ssh.Session, command *Command) (err error) {
	var (
		stdin          io.WriteCloser
		stdout, stderr io.Reader
	)

	if command.Stdin != nil {
		stdin, err = session.StdinPipe()
		if err != nil {
			return
		}

		go func() {
			_, err = io.Copy(stdin, command.Stdin)
			if err != nil {
				return
			}

			if err = stdin.Close(); err != nil {
				return
			}
		}()
	}

	if command.Stdout != nil {
		stdout, err = session.StdoutPipe()
		if err != nil {
			return
		}

		go func() {
			_, err = io.Copy(command.Stdout, stdout)
			if err != nil {
				return
			}
		}()
	}

	if command.Stderr != nil {
		stderr, err = session.StderrPipe()
		if err != nil {
			return
		}

		go func() {
			_, err = io.Copy(command.Stderr, stderr)
			if err != nil {
				return
			}
		}()
	}

	return
}

// setEnvVariables sets environment variables for the session.
func setEnvVariables(session *ssh.Session, env map[string]string) (err error) {
	for variable, value := range env {
		if err = session.Setenv(variable, value); err != nil {
			return
		}
	}

	return
}

// setPty sets up a pseudo-terminal for the session.
func setPty(session *ssh.Session) (err error) {
	sTerm := os.Getenv("TERM")
	if sTerm == "" {
		sTerm = "xterm-256color"
	}

	termmodes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		ssh.OPOST:         1,     // Enable output processing.
	}

	if err = session.RequestPty(sTerm, 40, 80, termmodes); err != nil {
		return
	}

	return
}
