package solidity

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type downloader struct {
	dst string

	// channel to sync concurrent solidity downloads
	downloaderLock sync.Mutex
	downloaderCh   map[string]chan struct{}
}

func (d *downloader) Exists(version string) bool {
	if _, err := os.Stat(s.Path(version)); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(err)
	}
}

func (d *downloader) download(version string) error {
	if s.Exists(version) {
		return nil
	}

	// check if we are already downloading it
	d.downloaderLock.Lock()
	ch, ok := d.downloaderCh[version]
	if !ok {
		ch = make(chan struct{})
		defer close(ch)

		d.downloaderCh[version] = ch
		d.downloaderLock.Unlock()

		// download the compiler since no one is doing it already
		if err := d.downloadImpl(version, d.dst); err != nil {
			return err
		}
	} else {
		<-ch
	}

	return nil
}

func (d *downloader) downloadImpl(version string, dst string) error {
	url := "https://github.com/ethereum/solidity/releases/download/v" + version + "/solc-static-linux"

	// check if the dst is correct
	exists := false
	fi, err := os.Stat(dst)
	if err == nil {
		switch mode := fi.Mode(); {
		case mode.IsDir():
			exists = true
		case mode.IsRegular():
			return fmt.Errorf("dst is a file")
		}
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat dst '%s': %v", dst, err)
		}
	}

	// create the destiny path if does not exists
	if !exists {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return fmt.Errorf("cannot create dst path: %v", err)
		}
	}

	// rename binary
	name := "solidity-" + version

	// tmp folder to download the binary
	tmpDir, err := ioutil.TempDir(dst, "solc-download-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, name)

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// make binary executable
	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	// move file to dst
	if err := os.Rename(path, filepath.Join(dst, name)); err != nil {
		return err
	}
	return nil
}

var solidityVersions = map[string]string{}
