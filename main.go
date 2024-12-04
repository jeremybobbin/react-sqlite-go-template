package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
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

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-t" {
		_, err := PrintStructs(os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error printing structs: %s\n", err.Error())
			os.Exit(1)
		}
		return
	}

	ctx := context.Background()
	muxer := http.NewServeMux()
	server := &http.Server{
		Addr: "localhost:8080",
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Handler: muxer,
	}

	database, err := sql.Open("sqlite3", "./sql.db")
	if err != nil {
		panic(err)
	}

	store := &Store{
		database: database,
	}

	muxer.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("public"))))
	muxer.Handle("/items/", http.StripPrefix("/items/", store))

	errs := make(chan error)
	go func() {
		errs <- server.ListenAndServe()
		close(errs)
	}()
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP)
	<-signals
	if err := server.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "error closing server: %s\n", err.Error())
	}
}
