package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"strconv"
	"time"
)

type Store struct {
	database *sql.DB
}

func (store *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	ctx, cancel := context.WithCancel(r.Context())

	var item Item
	var id int
	var err error

	switch r.Method {
	case "PATCH", "DELETE":
		id, err = strconv.Atoi(path)
		if err != nil {
			http.Error(w, "unable to parse ID", http.StatusBadRequest)
			return
		}
	}

	switch r.Method {
	case "POST", "PATCH":
		switch r.Header.Get("Content-Type") {
		case "application/json":
			dec := json.NewDecoder(r.Body)
			if err := dec.Decode(&item); err != nil {
				http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
				return
			}
		case "text/plain":
			if err := TextDecode(r.Body, &item); err != nil {
				http.Error(w, "invalid text: "+err.Error(), http.StatusBadRequest)
				return
			}
		default: // plain text
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
	}

	switch r.Method {
	case "GET":
		var err error
		ch := make(chan struct {
			int
			Item
		})
		go func() {
			if err = GetItems(ctx, store.database, ch); err != nil {
				cancel()
			}
			close(ch)
		}()

		switch accepted := r.Header.Get("Accept"); accepted {
		case "application/json":
			items := make(map[int]Item)
			for pair := range ch {
				items[pair.int] = pair.Item
			}
			if err != nil {
				http.Error(w, "error fetching items: "+err.Error(), http.StatusInternalServerError)
			}
			data, err := json.Marshal(items)
			if err != nil {
				http.Error(w, "error encoding items: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", accepted)
			w.Write(data)
			return
		case "text/plain", "*/*", "":
			w.Header().Set("Content-Type", "text/plain")
			for pair := range ch {
				id := pair.int
				item := pair.Item

				fmt.Fprintf(w, "%-12s %d\n", "ID", id)
				fmt.Fprintf(w, "%-12s %s\n", "Description", item.Description)
				if item.Done {
					fmt.Fprintf(w, "%-12s true\n", "Done")
				} else {
					fmt.Fprintf(w, "%-12s false\n", "Done")
				}
				fmt.Fprintf(w, "%-12s %s\n\n", "Time", item.Time.Format(time.RFC3339))
			}
			if err != nil {
				http.Error(w, "error fetching items", http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "unsupported type "+accepted, http.StatusUnsupportedMediaType)
			return
		}
	case "POST":
		_, err := store.database.ExecContext(ctx, "INSERT INTO items (description, time, done) VALUES (?, ?, ?)", item.Description, item.Time, item.Done)
		if err != nil {
			http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	case "PATCH":
		res, err := store.database.ExecContext(ctx, "UPDATE items SET description = ?, done = ? WHERE id = ?", item.Description, item.Done, id)
		if err != nil {
			http.Error(w, "store.database error: "+err.Error(), http.StatusInternalServerError)
		} else if rows, err := res.RowsAffected(); err != nil {
			http.Error(w, "store.database error: "+err.Error(), http.StatusNotFound)
		} else if rows == 0 {
			http.Error(w, "store.database error: item not found", http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	case "DELETE":
		res, err := store.database.ExecContext(ctx, "DELETE FROM items WHERE items.id = ?", id)
		if err != nil {
			http.Error(w, "store.database error: "+err.Error(), http.StatusInternalServerError)
		} else if rows, err := res.RowsAffected(); err != nil {
			http.Error(w, "store.database error: "+err.Error(), http.StatusNotFound)
		} else if rows == 0 {
			http.Error(w, "store.database error: item not found", http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func GetItems(ctx context.Context, database *sql.DB, ch chan struct {
	int
	Item
}) error {
	rows, err := database.QueryContext(ctx, "SELECT id, description, time, done FROM items")
	if err != nil {
		return err
	}

	var e struct {
		int
		Item
	}

loop:
	for rows.Next() {
		err = rows.Scan(&e.int, &e.Item.Description, &e.Item.Time, &e.Item.Done)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			break loop
		case ch <- e:
		}
	}

	return rows.Close()
}
