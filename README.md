# hqgossh

[![license](https://img.shields.io/badge/license-MIT-gray.svg?color=0040FF)](https://github.com/hueristiq/hqgossh/blob/master/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-0040ff.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hqgossh.svg?style=flat&color=0040ff)](https://github.com/hueristiq/hqgossh/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hqgossh.svg?style=flat&color=0040ff)](https://github.com/hueristiq/hqgossh/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-0040ff.svg)](https://github.com/hueristiq/hqgossh/blob/master/CONTRIBUTING.md)

A [Go(Golang)](https://golang.org/) package to provide a simple abstraction around [ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) and [sftp](https://pkg.go.dev/github.com/pkg/sftp) packages.

## Resources

* [Features](#features)
* [Usage](#usage)
	* [Authentication with password](#authentication-with-password)
	* [Authentication with key with passphrase](#authentication-with-key-with-passphrase)
	* [Authentication with Key without passphrase](#authentication-with-key-without-passphrase)
	* [Run remote commands](#run-remote-commands)
	* [Attach interactive shell](#attach-interactive-shell)
	* [Files upload and download](#files-upload-and-download)
		* [Upload Local File/Directory to Remote](#upload-local-filedirectory-to-remote)
		* [Download Remote File/Directory to Local](#download-remote-filedirectory-to-local)
* [Contribution](#contribution)

## Features

- [x] Authentication with password.
- [x] Authentication with keys with passphrase.
- [x] Authentication with keys without passphrase.
- [x] Supports running remote commands.
- [x] Supports getting an interactive shell.
- [x] Supports files upload and download.

## Installation

```bash
go get -v -u github.com/hueristiq/hqgossh
```

## Usage

### Authentication with password

```go
auth, err := authentication.Password("Password")
if err != nil {
	log.Fatal(err)
}

client, err := ssh.New(&ssh.Configuration{
	Host:            "xxx.xxx.xxx.xxx",
	Port:            22,
	User:            "some-user",
	Authentication:  auth,
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
})
if err != nil {
    log.Println(err)
}

defer client.Close()
```

### Authentication with key with passphrase

```go
auth, err := authentication.KeyWithPassphrase(privateKey, "Passphrase")
if err != nil {
	log.Fatal(err)
}

client, err := ssh.New(&ssh.Configuration{
	Host:            "xxx.xxx.xxx.xxx",
	Port:            22,
	User:            "some-user",
	Authentication:  auth,
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
})
if err != nil {
	log.Println(err)
}

defer client.Close()
```

### Authentication with key without passphrase

```go
auth, err := authentication.Key(privateKey)
if err != nil {
	log.Fatal(err)
}

client, err := ssh.New(&ssh.Configuration{
	Host:            "xxx.xxx.xxx.xxx",
	Port:            22,
	User:            "some-user",
	Authentication:  auth,
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
})
if err != nil {
	log.Println(err)
}

defer client.Close()
```

### Run remote commands

```go
if err = client.Run(&ssh.Command{
	CMD:    "echo ${LC_TEST}",
	ENV:    map[string]string{"LC_TEST":"working"},
	Stdin:  os.Stdin,
	Stdout: os.Stdout,
	Stderr: os.Stderr,
}); err != nil {
	log.Fatal(err)
}
```

### Attach interactive shell

```go
if err = client.Shell(); err != nil {
	log.Println(err)
}
```

### Files upload and download

#### Upload Local File/Directory to Remote

```go
if err := client.Upload("/path/to/local/file", "/path/to/remote/file"); err != nil {
	log.Println(err)
}
```

#### Download Remote File/Directory to Local

```go
if err := client.Download("/path/to/remote/file", "/path/to/local/file"); err != nil {
	log.Println(err)
}
```

## Contributing

[Issues](https://github.com/hueristiq/hqgossh/issues) and [Pull Requests](https://github.com/hueristiq/hqgossh/pulls) are welcome! **Check out the [contribution guidelines](./CONTRIBUTING.md)**.

## Licensing

This tool is distributed under the [MIT license](https://github.com/hueristiq/hqgossh/blob/master/LICENSE).