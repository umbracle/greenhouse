package solidity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type Solidity struct {
	// Destination folder for solidity compiler downloads
	Dst string

	// channel to sync concurrent solidity downloads
	downloaderLock sync.Mutex
	downloaderCh   map[string]chan struct{}
}

func NewSolidity(dir string) *Solidity {
	return &Solidity{
		Dst:          dir,
		downloaderCh: map[string]chan struct{}{},
	}
}

func (s *Solidity) download(version string) error {
	if s.Exists(version) {
		return nil
	}

	// check if we are already downloading it
	s.downloaderLock.Lock()
	ch, ok := s.downloaderCh[version]
	if !ok {
		ch = make(chan struct{})
		defer close(ch)

		s.downloaderCh[version] = ch
		s.downloaderLock.Unlock()

		// download the compiler since no one is doing it already
		if err := downloadSolidity(version, s.Dst); err != nil {
			return err
		}
	} else {
		<-ch
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

	//fmt.Println("-- args --")
	//fmt.Println(args)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(path, args...)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to compile: %s", stderr.String())
	}

	fmt.Println(stdout.String())

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
