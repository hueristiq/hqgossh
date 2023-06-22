package authentication

import (
	"golang.org/x/crypto/ssh"
)

type Authentication []ssh.AuthMethod

func Password(password string) (authentication Authentication) {
	authentication = Authentication{
		ssh.Password(password),
	}

	return
}

func KeyWithPassphrase(privateKey, passphrase string) (authentication Authentication, err error) {
	var signer ssh.Signer

	signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passphrase))
	if err != nil {
		return
	}

	authentication = Authentication{
		ssh.PublicKeys(signer),
	}

	return
}

func KeyWithoutPassphrase(privateKey string) (authentication Authentication, err error) {
	var signer ssh.Signer

	signer, err = ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return
	}

	authentication = Authentication{
		ssh.PublicKeys(signer),
	}

	return
}
