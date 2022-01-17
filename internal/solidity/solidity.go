package solidity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type Solidity struct {
	// Destination folder for solidity compiler downloads
	Dst string
}

func NewSolidity(dir string) *Solidity {
	return &Solidity{
		Dst: dir,
	}
}

func (s *Solidity) download(version string) error {
	if s.Exists(version) {
		return nil
	}
	// download the compiler since no one is doing it already
	fmt.Printf("Downloading solidity %s...\n", version)
	if err := downloadSolidity(version, s.Dst); err != nil {
		return err
	}
	return nil
}

func (s *Solidity) Compile(input *Input) (*Output, error) {
	version := input.Version
	files := input.Files

	if err := s.download(version); err != nil {
		return nil, err
	}

	path := s.Path(version)

	args := []string{
		"--combined-json",
		"bin,bin-runtime,srcmap-runtime,abi,srcmap,ast",
	}
	if len(input.Remappings) != 0 {
		for k, v := range input.Remappings {
			args = append(args, k+"="+v)
		}
	}

	if len(files) != 0 {
		args = append(args, files...)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(path, args...)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to compile: %s", stderr.String())
	}
	var output *Output
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Solidity) Path(version string) string {
	return filepath.Join(s.Dst, "solidity-"+version)
}

func (s *Solidity) Exists(version string) bool {
	if _, err := os.Stat(s.Path(version)); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(err)
	}
}

func downloadSolidity(version string, dst string) error {
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
