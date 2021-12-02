package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

// ErrCacheFileNotExists The expected cache-file does not exist.
var ErrCacheFileNotExists = errors.New("cache file not exists")

type ResponseCache struct {
	notEmpty bool
	key      string
	filename string
	next     *ResponseCache
	prev     *ResponseCache
	mu       sync.RWMutex
}

func (rc *ResponseCache) NotEmpty() bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	return rc.notEmpty
}

func (rc *ResponseCache) Save(headers http.Header, body []byte) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	f, err := os.Create(rc.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// write headers to file
	for header, values := range headers {
		for _, val := range values {
			if _, err := f.WriteString(header + ":" + val + "\n"); err != nil {
				return err
			}
		}
	}
	// write head-body delimiter
	if _, err := f.WriteString("\n"); err != nil {
		return err
	}
	if _, err := f.Write(body); err != nil {
		return err
	}
	rc.notEmpty = true
	return nil
}

func (rc *ResponseCache) WriteTo(w http.ResponseWriter) error {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if _, err := os.Stat(rc.filename); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrCacheFileNotExists, rc.filename)
	}

	f, err := os.Open(rc.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	// read headers
	for {
		line, err := r.ReadString('\n')
		if errors.Is(err, io.EOF) || line == "\n" {
			break
		}
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		heads := strings.SplitN(line, ":", 2)
		if len(heads) != 2 {
			return errors.New("malformed cache file (headers)")
		}
		w.Header().Add(heads[0], heads[1])
	}

	// read body
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if errors.Is(err, io.EOF) || n == 0 {
			break
		}
		if err != nil {
			return err
		}
		if _, err := w.Write(buf[0:n]); err != nil {
			return err
		}
	}

	return nil
}

func (rc *ResponseCache) Remove() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.filename != "" {
		return os.Remove(rc.filename)
	}
	return nil
}
