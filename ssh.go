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
	"golang.org/x/crypto/ssh/terminal"
)

// Client represents a client consisting of an SSH client and an SFTP session
type Client struct {
	SSH  *ssh.Client
	SFTP *sftp.Client
}

// Options represents options used in creating a Client
type Options struct {
	Host            string
	Port            int
	User            string
	Authentication  authentication.Authentication
	Timeout         int
	HostKeyCallback ssh.HostKeyCallback
}

const (
	DefaultTimeout = 30
)

// New create Client
func New(options *Options) (client *Client, err error) {
	client = &Client{
		SSH:  &ssh.Client{},
		SFTP: &sftp.Client{},
	}

	if options.Timeout <= 0 {
		options.Timeout = DefaultTimeout
	}

	server := net.JoinHostPort(options.Host, fmt.Sprint(options.Port))
	config := &ssh.ClientConfig{
		User:            options.User,
		Auth:            options.Authentication,
		Timeout:         time.Duration(options.Timeout) * time.Second,
		HostKeyCallback: options.HostKeyCallback,
	}

	retry, retryDelay, maxRetries := 1, 5, 10

CREATE_CLIENT:
	if client.SSH, err = ssh.Dial("tcp", server, config); err != nil {
		if retry <= maxRetries {
			time.Sleep(time.Duration(retryDelay) * time.Second)

			retry++

			goto CREATE_CLIENT
		}

		return
	}

	retry = 1

	if client.SFTP, err = sftp.NewClient(client.SSH); err != nil {
		if retry <= maxRetries {
			time.Sleep(time.Duration(retryDelay) * time.Second)

			retry++

			goto CREATE_CLIENT
		}

		return
	}

	return
}

// Run runs cmd on the remote host.
func (client *Client) Run(command *Command) (err error) {
	var (
		session        *ssh.Session
		stdin          io.WriteCloser
		stdout, stderr io.Reader
	)

	session, err = client.SSH.NewSession()
	if err != nil {
		return
	}

	defer session.Close()

	if command.Stdin != nil {
		stdin, err = session.StdinPipe()
		if err != nil {
			return
		}

		go io.Copy(stdin, command.Stdin)
	}

	if command.Stdout != nil {
		stdout, err = session.StdoutPipe()
		if err != nil {
			return
		}

		go io.Copy(command.Stdout, stdout)
	}

	if command.Stderr != nil {
		stderr, err = session.StderrPipe()
		if err != nil {
			return
		}

		go io.Copy(command.Stderr, stderr)
	}

	for variable, value := range command.ENV {
		if err = session.Setenv(variable, value); err != nil {
			return
		}
	}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}

	termmodes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		ssh.OPOST:         1,     // Enable output processing.
	}

	if err = session.RequestPty(term, 40, 80, termmodes); err != nil {
		return
	}

	if err = session.Run(command.CMD); err != nil {
		return
	}

	return
}

// Shell starts a login shell on the remote host.
func (client *Client) Shell() (err error) {
	var (
		session *ssh.Session
	)

	session, err = client.SSH.NewSession()
	if err != nil {
		return
	}

	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}

	termmodes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		ssh.OPOST:         1,     // Enable output processing.
	}

	fd := int(os.Stdin.Fd())

	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return
	}

	defer terminal.Restore(fd, state)

	width, height, err := terminal.GetSize(fd)
	if err != nil {
		return
	}

	if err = session.RequestPty(term, height, width, termmodes); err != nil {
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

// Close closes  SFTP session and the underlying network connection
func (client *Client) Close() (err error) {
	if client == nil {
		return
	}

	if client.SFTP != nil {
		if err = client.SFTP.Close(); err != nil {
			return
		}
	}

	if client.SSH != nil {
		if err = client.SSH.Close(); err != nil {
			return
		}
	}

	return
}
