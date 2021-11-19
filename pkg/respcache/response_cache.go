package respcache

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var ErrCacheFileNotExists = errors.New("cache file not exists")

type ResponseCache struct {
	status   int
	header   http.Header
	key      string
	filename string
	file     *os.File
	next     *ResponseCache
	prev     *ResponseCache
	mu       sync.RWMutex
}

func (rc *ResponseCache) WriteHeader(statusCode int) {
	rc.status = statusCode
}

func (rc *ResponseCache) GetStatus() int {
	return rc.status
}

func (rc *ResponseCache) Header() http.Header {
	return rc.header
}

func (rc *ResponseCache) SetHeader(h http.Header) {
	rc.header = h
}

func (rc *ResponseCache) Write(bytes []byte) (int, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// if start writing now
	if rc.file == nil {
		f, err := os.Create(rc.filename)
		if err != nil {
			return 0, err
		}
		// write status
		if _, err := f.WriteString(strconv.Itoa(rc.status) + "\n"); err != nil {
			return 0, err
		}
		// write headers to file
		for header, values := range rc.header {
			for _, val := range values {
				if _, err := f.WriteString(header + ":" + val + "\n"); err != nil {
					return 0, err
				}
			}
		}
		// write head-body delimiter
		if _, err := f.WriteString("\n"); err != nil {
			return 0, err
		}
		rc.file = f
	}
	// write body
	return rc.file.Write(bytes)
}

func (rc *ResponseCache) Close() error {
	// while someone writes to file, it cannot be closed
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.file != nil {
		if err := rc.file.Close(); err != nil {
			return err
		}
		rc.file = nil
	}
	return nil
}

func (rc *ResponseCache) Read(w http.ResponseWriter) error {
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

	//read status
	line, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	line = strings.TrimSpace(line)
	status, err := strconv.Atoi(line)
	if err != nil {
		return errors.New("malformed cache file (status)")
	}
	w.WriteHeader(status)
	// read headers
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF || line == "\n" {
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
		if err == io.EOF || n == 0 {
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

func (rc *ResponseCache) Remove() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.filename != "" {
		os.Remove(rc.filename)
	}
}
