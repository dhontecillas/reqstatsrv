package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dhontecillas/reqstatsrv/config"
	"github.com/dhontecillas/reqstatsrv/endpoint"
)

func main() {
	fmt.Println("starting request stats server...")

	cfg := config.DefaultConfig()
	if len(os.Args) > 1 {
		fmt.Printf("loading from config file %s\n", os.Args[1])
		fname := os.Args[1]
		b, err := os.ReadFile(fname)
		if err != nil {
			fmt.Printf("cannot read config file %s: %s\n", fname, err.Error())
			return
		}
		err = json.Unmarshal(b, &cfg)
		if err != nil {
			fmt.Printf("cannot parse config file %s: %s\n", fname, err.Error())
			return
		}
	}
	fmt.Printf("config:\n%#v\n", cfg)
	launchServer(cfg)
}

func launchServer(cfg *config.Config) {
	mux := http.NewServeMux()
	for _, e := range cfg.Endpoints {
		endpoint.Bind(mux, &e)
	}

	srv := &http.Server{
		Addr:    cfg.Addr(),
		Handler: mux,
	}
	sigChan := make(chan os.Signal, 1)
	go signalHandler(sigChan, srv)
	if err := srv.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			fmt.Printf("error %s\nSHUTTING DOWN", err.Error())
		}
	}
}

func signalHandler(sc chan os.Signal, srv *http.Server) {
	signal.Notify(sc, os.Interrupt)
	<-sc
	fmt.Printf("\nshutdown signal received\n")
	shutdownServer(srv)
}

func shutdownServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
