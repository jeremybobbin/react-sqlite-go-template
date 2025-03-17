package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Item struct {
	Description string
	Done        bool
	Time        time.Time
}

var (
	Address = "0.0.0.0"
	Port    = 8080
)

func init() {
	flag.StringVar(&Address, "a", Address, "address for HTTP server")
	flag.IntVar(&Port, "p", Port, "port for HTTP server")
	flag.Parse()

	if Port < 1 || Port > 65535 {
		fmt.Fprintf(os.Stderr, "error - invalid port number: %d\n", Port)
		os.Exit(1)
	}

	if ip := net.ParseIP(Address); ip == nil {
		fmt.Fprintf(os.Stderr, "error - invalid address: %s\n", Address)
		os.Exit(1)
	}
}

func main() {
	muxer := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", Address, Port),
		Handler: muxer,
	}

	database, err := sql.Open("sqlite3", "./sql.db")
	if err != nil {
		panic(err)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		err = database.PingContext(ctx)
		if err != nil {
			panic(err)
		}
		cancel()
	}

	store := &Store{
		database: database,
	}

	muxer.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))
	muxer.Handle("/items/", http.StripPrefix("/items/", store))

	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt, syscall.SIGHUP)
		<-ch
		server.Close()
	}()

	fmt.Fprintf(os.Stderr, "serving at %s:%d\n", Address, Port)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error closing server: %s\n", err.Error())
	}
	fmt.Println("closed server - closing db")
	if err = database.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "error closing database: %s\n", err.Error())
	}
}
