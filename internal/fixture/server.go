package fixture

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/kimjooyoon/gooo-content-addressed-proof-reuse/internal/reuse"
)

type Server struct {
	bundle *reuse.Bundle
	server *httptest.Server
	mu     sync.Mutex
	stats  reuse.FetchStats
}

func NewServer(bundle reuse.Bundle) *Server {
	fixtureServer := &Server{bundle: &bundle}
	fixtureServer.server = httptest.NewServer(http.HandlerFunc(fixtureServer.handle))
	return fixtureServer
}

func (server *Server) Close() { server.server.Close() }

func (server *Server) URL() string { return server.server.URL }

func (server *Server) ResetStats() {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.stats = reuse.FetchStats{}
}

func (server *Server) Stats() reuse.FetchStats {
	server.mu.Lock()
	defer server.mu.Unlock()
	return server.stats
}

func (server *Server) Fetch(lock reuse.Lock) ([]byte, error) {
	client := http.Client{Timeout: time.Second}
	response, err := client.Get(server.server.URL + "/v1/locks/" + url.PathEscape(lock.Coordinate))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fixture server returned %d for %s", response.StatusCode, lock.Coordinate)
	}
	return body, nil
}

func (server *Server) handle(response http.ResponseWriter, request *http.Request) {
	prefix := "/v1/locks/"
	if !strings.HasPrefix(request.URL.Path, prefix) {
		http.NotFound(response, request)
		return
	}
	coordinate, err := url.PathUnescape(strings.TrimPrefix(request.URL.Path, prefix))
	if err != nil || !server.hasCoordinate(coordinate) {
		http.NotFound(response, request)
		return
	}
	body := []byte("gooo/content-addressed-proof-reuse/fixture/v1|" + coordinate)
	server.mu.Lock()
	server.stats.Requests++
	server.stats.BytesRead += int64(len(body))
	server.stats.BytesDownloaded += int64(len(body))
	server.mu.Unlock()
	response.Header().Set("Content-Type", "application/octet-stream")
	_, _ = response.Write(body)
}

func (server *Server) hasCoordinate(coordinate string) bool {
	for _, lock := range server.bundle.ParentReceipt.Locks {
		if lock.Coordinate == coordinate {
			return true
		}
	}
	return false
}
