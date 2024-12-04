package main

import (
	"context"
	"database/sql"
	"fmt"
	//"net"
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
	muxer := http.NewServeMux()
	server := &http.Server{
		Addr:    "localhost:8080",
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

	fmt.Println("serving")
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error closing server: %s\n", err.Error())
	}
	fmt.Println("closed server - closing db")
	if err = database.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "error closing database: %s\n", err.Error())
	}
}
