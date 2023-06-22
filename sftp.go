package hqgossh

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	confirmSourceType = true

	ErrSourceNotFound    = errors.New("source not found")
	ErrSourceIsDirectory = errors.New("source is directory")
	ErrSourceIsFile      = errors.New("source is file")
)

// Upload transfers local directories and files to remote host.
// This works by calling UploadDirectory or UploadFile depending on the source.
func (client *Client) Upload(SRC, DEST string) (err error) {
	confirmSourceType = false

	stat, err := os.Stat(SRC)
	if os.IsNotExist(err) {
		return ErrSourceNotFound
	} else if err != nil && !os.IsNotExist(err) {
		return
	}

	if stat.IsDir() {
		if err = client.UploadDirectory(SRC, DEST); err != nil {
			return
		}
	} else {
		if err = client.UploadFile(SRC, DEST); err != nil {
			return
		}
	}

	return
}

// UploadDirectory transfers local directories and contained files recursively to remote host
func (client *Client) UploadDirectory(SRC, DEST string) (err error) {
	if confirmSourceType {
		var stat os.FileInfo

		stat, err = os.Stat(SRC)
		if os.IsNotExist(err) {
			return ErrSourceNotFound
		} else if err != nil && !os.IsNotExist(err) {
			return
		}

		if !stat.IsDir() {
			return ErrSourceIsFile
		}
	}

	if err = filepath.Walk(SRC, func(fileSRC string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileDEST := filepath.Join(DEST, strings.TrimPrefix(fileSRC, SRC))

			if err := client.UploadFile(fileSRC, fileDEST); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return
	}

	return
}

// UploadFile transfers local files to remote host
func (client *Client) UploadFile(SRC, DEST string) (err error) {
	if confirmSourceType {
		var stat os.FileInfo

		stat, err = os.Stat(SRC)
		if os.IsNotExist(err) {
			return ErrSourceNotFound
		} else if err != nil && !os.IsNotExist(err) {
			return
		}

		if stat.IsDir() {
			return ErrSourceIsDirectory
		}
	}

	directory := filepath.Dir(DEST)

	_, err = client.SFTP.Stat(directory)
	if os.IsNotExist(err) {
		if err = client.SFTP.MkdirAll(directory); err != nil {
			return
		}
	} else if err != nil && !os.IsNotExist(err) {
		return
	}

	SRCFile, err := os.Open(SRC)
	if err != nil {
		return
	}

	defer SRCFile.Close()

	DESTFile, err := client.SFTP.OpenFile(DEST, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		return
	}

	defer DESTFile.Close()

	if _, err = io.Copy(DESTFile, SRCFile); err != nil {
		return
	}

	return
}

// Download transfers remote directories and files to local host.
// This works by calling DownloadDirectory or DownloadFile depending on the source.
func (client *Client) Download(SRC, DEST string) (err error) {
	confirmSourceType = false

	stat, err := client.SFTP.Stat(SRC)
	if os.IsNotExist(err) {
		return ErrSourceNotFound
	} else if err != nil && !os.IsNotExist(err) {
		return
	}

	if stat.IsDir() {
		if err = client.DownloadDirectory(SRC, DEST); err != nil {
			return
		}
	} else {
		if err = client.DownloadFile(SRC, DEST); err != nil {
			return
		}
	}

	return
}

// DownloadDirectory transfers remote directories and contained files recursively to local host
func (client *Client) DownloadDirectory(SRC, DEST string) (err error) {
	if confirmSourceType {
		var stat os.FileInfo

		stat, err = client.SFTP.Stat(SRC)
		if os.IsNotExist(err) {
			return ErrSourceNotFound
		} else if err != nil && !os.IsNotExist(err) {
			return
		}

		if !stat.IsDir() {
			return ErrSourceIsFile
		}
	}

	walker := client.SFTP.Walk(SRC)

	for walker.Step() {
		if err = walker.Err(); err != nil {
			return
		}

		if !walker.Stat().IsDir() {
			fileDEST := filepath.Join(DEST, strings.TrimPrefix(walker.Path(), SRC))

			if err = client.DownloadFile(walker.Path(), fileDEST); err != nil {
				return
			}
		}
	}

	return
}

// DownloadFile transfers remote files to local host
func (client *Client) DownloadFile(SRC, DEST string) (err error) {
	if confirmSourceType {
		var stat os.FileInfo

		stat, err = client.SFTP.Stat(SRC)
		if os.IsNotExist(err) {
			return ErrSourceNotFound
		} else if err != nil && !os.IsNotExist(err) {
			return
		}

		if stat.IsDir() {
			return ErrSourceIsDirectory
		}
	}

	directory := filepath.Dir(DEST)

	if _, err = os.Stat(directory); os.IsNotExist(err) {
		if err = os.MkdirAll(directory, os.ModePerm); err != nil {
			return
		}
	}

	SRCFile, err := client.SFTP.OpenFile(SRC, (os.O_RDONLY))
	if err != nil {
		return
	}

	defer SRCFile.Close()

	DESTFile, err := os.Create(DEST)
	if err != nil {
		return
	}

	defer DESTFile.Close()

	if _, err = io.Copy(DESTFile, SRCFile); err != nil {
		return
	}

	return
}
