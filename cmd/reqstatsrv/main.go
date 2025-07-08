package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/dhontecillas/reqstatsrv/server"
)

func main() {
	fmt.Println("starting request stats server...")

	// TODO: if no args are provided, then we can use some
	// embedded default config
	// remove the DefaultConfig
	// cfg := config.DefaultConfig()
	confFile := ""
	if len(os.Args) > 1 {
		confFile = os.Args[1]
	}

	srv, err := server.NewHTTPServer(context.Background(), confFile)
	if err != nil {
		fmt.Printf("cannot start http server: %s\n", err.Error())
		return
	}

	go signalHandler(srv)
	// fmt.Printf("config:\n%#v\n", cfg)
	srv.Run()
}

func signalHandler(srv *server.HTTPServer) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Printf("\nshutdown signal received\n")
	srv.Shutdown()
}
