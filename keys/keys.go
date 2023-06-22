package keys

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

func Read(path string) (priv string, pub string, err error) {
	var (
		stats           os.FileInfo
		privBuf, pubBUf []byte
	)

	stats, err = os.Stat(path)
	if err != nil {
		err = fmt.Errorf("failed reading private key: %s", err)

		return
	}

	if stats.IsDir() {
		err = fmt.Errorf("failed reading private key: %s is a directory!", path)

		return
	}

	privBuf, err = ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("failed reading private key: %s", err)

		return
	}

	priv = string(privBuf)

	pubBUf, err = ioutil.ReadFile(path + ".pub")
	if err != nil {
		err = fmt.Errorf("failed reading public key: %s", err)

		return
	}

	pub = string(pubBUf)

	return
}

func Generate() (priv string, pub string, err error) {
	var (
		privateKey *rsa.PrivateKey
		publicKey  ssh.PublicKey
		privBuf    bytes.Buffer
		pubBUf     []byte
	)

	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("failed generating private key: %s", err)

		return
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err = pem.Encode(&privBuf, privateKeyPEM); err != nil {
		err = fmt.Errorf("failed encoding private key: %s", err)

		return
	}

	priv = privBuf.String()

	publicKey, err = ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		err = fmt.Errorf("failed generating public key: %s", err)

		return
	}

	pubBUf = ssh.MarshalAuthorizedKey(publicKey)
	pub = string(pubBUf)

	return
}

func ReadOrGenerate(path string) (priv string, pub string, err error) {
	pub, priv, err = Read(path)
	if err != nil {
		goto GENERATE
	} else {
		return
	}

GENERATE:
	pub, priv, err = Generate()
	if err != nil {
		err = fmt.Errorf("failed generating keys: %s", err)

		return
	}

	if err = Write(path, pub, priv); err != nil {
		err = fmt.Errorf("failed writing keys: %s", err)

		return
	}

	return
}

func Write(path, pub, priv string) (err error) {
	directory := filepath.Dir(path)

	if _, err = os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			if directory != "" {
				if err = os.MkdirAll(directory, os.ModePerm); err != nil {
					return
				}
			}
		} else {
			return
		}
	}

	if err = ioutil.WriteFile(path, []byte(priv), 0600); err != nil {
		return
	}

	if err = ioutil.WriteFile(path+".pub", []byte(pub), 0644); err != nil {
		return
	}

	return
}
