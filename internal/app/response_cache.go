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
	key      string
	filepath string
	next     *ResponseCache
	prev     *ResponseCache
	mu       sync.RWMutex
}

func (rc *ResponseCache) save(headers http.Header, body []byte) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	f, err := os.Create(rc.filepath)
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

	return nil
}

func (rc *ResponseCache) writeTo(w http.ResponseWriter) error {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if _, err := os.Stat(rc.filepath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrCacheFileNotExists, rc.filepath)
	}

	f, err := os.Open(rc.filepath)
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
			fmt.Println(line)
			return errors.New("malformed cache file (headers)")
		}
		w.Header().Add(heads[0], heads[1])
	}

	w.Header().Add("X-Minipic-Cache", "HIT")

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

func (rc *ResponseCache) remove() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.filepath != "" {
		return os.Remove(rc.filepath)
	}
	return nil
}
