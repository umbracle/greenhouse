package solidity

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// SolidityVersionManager is a manager to download solidity versions
type SolidityVersionManager struct {
	Dst string
}

func NewSolidityVersionManager(dir string) (*SolidityVersionManager, error) {
	if dir == "" {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		dir = filepath.Join(dirname, ".gosolc")
	}
	s := &SolidityVersionManager{
		Dst: dir,
	}
	return s, nil
}

func (s *SolidityVersionManager) Path(version string) string {
	return filepath.Join(s.Dst, "solidity-"+version)
}

func (s *SolidityVersionManager) Exists(version string) bool {
	if _, err := os.Stat(s.Path(version)); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(err)
	}
}

func (s *SolidityVersionManager) Download(version string) error {
	url := "https://github.com/ethereum/solidity/releases/download/v" + version + "/solc-static-linux"

	// check if the dst is correct
	exists := false
	fi, err := os.Stat(s.Dst)
	if err == nil {
		switch mode := fi.Mode(); {
		case mode.IsDir():
			exists = true
		case mode.IsRegular():
			return fmt.Errorf("dst is a file")
		}
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat dst '%s': %v", s.Dst, err)
		}
	}

	// create the destiny path if does not exists
	if !exists {
		if err := os.MkdirAll(s.Dst, 0755); err != nil {
			return fmt.Errorf("cannot create dst path: %v", err)
		}
	}

	// rename binary
	name := "solidity-" + version

	// tmp folder to download the binary
	tmpDir, err := ioutil.TempDir(s.Dst, "solc-download-")
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
	if err := os.Rename(path, filepath.Join(s.Dst, name)); err != nil {
		return err
	}
	return nil
}

func (s *SolidityVersionManager) Compile(version string, input *Input) (*Output, error) {
	return nil, nil
}
