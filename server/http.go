package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dhontecillas/reqstatsrv/config"
	"github.com/dhontecillas/reqstatsrv/endpoint"

	"github.com/fsnotify/fsnotify"
)

var (
	ErrCannotReadConfigFile  = errors.New("cannot read config file")
	ErrCannotParseConfigFile = errors.New("cannot parse config file")
)

type HTTPServer struct {
	ctx        context.Context
	configFile string
	watcher    *fsnotify.Watcher
	srv        *http.Server
}

func setupWatcher(configFile string) *fsnotify.Watcher {
	var err error
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("cannot autoreload: %s\n", err.Error())
		return nil
	}
	// we watch the directory
	absCfgFile, err := filepath.Abs(configFile)
	if err != nil {
		fmt.Printf("cannot get absolute path for config file")
		watcher.Close()
		return nil
	}
	watcher.Add(filepath.Dir(absCfgFile))
	return watcher
}

func NewHTTPServer(ctx context.Context, configFile string) (*HTTPServer, error) {
	cfg, err := LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	var watcher *fsnotify.Watcher
	if cfg.Autoreload {
		watcher = setupWatcher(configFile)
	}

	srv, err := newServer(cfg)
	if err != nil {
		return nil, err
	}

	h := &HTTPServer{
		ctx:        ctx,
		configFile: configFile,
		watcher:    watcher,
		srv:        srv,
	}

	if watcher != nil {
		go h.backgroundWatch()
	}
	return h, nil
}

func LoadConfig(source string) (*config.Config, error) {
	var cfg config.Config
	b, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %s", ErrCannotReadConfigFile, source, err.Error())
	}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %s", ErrCannotParseConfigFile, source, err.Error())
	}
	return &cfg, nil
}

func (s *HTTPServer) backgroundWatch() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			// fmt.Printf("\nEVENT\n%s\n", event.String())
			if !ok {
				fmt.Printf("\ncannot continue watching the config file\n")
				return
			}
			if filepath.Base(event.Name) != filepath.Base(s.configFile) {
				continue
			}
			if !event.Has(fsnotify.Write) {
				continue
			}
			time.Sleep(100 * time.Millisecond)
			if err := s.reload(); err != nil {
				fmt.Printf("cannot launch new config %q: %s\nOp: %d -> %q",
					s.configFile, err.Error(), event.Op, event.Name)
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				fmt.Printf("\ncannot read error: wont watch config file anymore\n")
				return
			}
			fmt.Printf("watch error: %s\n", err.Error())
		case <-s.ctx.Done():
			fmt.Println("watch: shutting down")
		}
	}
}

func (s *HTTPServer) reload() error {
	cfg, cfgErr := LoadConfig(s.configFile)
	if cfgErr != nil {
		return cfgErr
	}

	srv, srvErr := newServer(cfg)
	if srvErr != nil {
		return srvErr
	}

	oldSrv := s.srv
	s.srv = srv

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	oldSrv.Shutdown(ctx)
	cancel()
	return nil
}

func (s *HTTPServer) Shutdown() {
	srv := s.srv
	s.srv = nil
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	srv.Shutdown(ctx)
	cancel()
}

func newServer(cfg *config.Config) (*http.Server, error) {
	mux := http.NewServeMux()
	for _, e := range cfg.Endpoints {
		endpoint.Bind(mux, &e)
	}

	return &http.Server{
		Addr:    cfg.Addr(),
		Handler: mux,
	}, nil
}

func (s *HTTPServer) Run() {
	for s.srv != nil {
		curSrv := s.srv
		if err := curSrv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("error %s\nSHUTTING DOWN\n", err.Error())
			} else if s.srv != nil {
				fmt.Printf("relaunching server\n")
			}
		}
	}
	fmt.Printf("cleanly shutting down\n")
}
