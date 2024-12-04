package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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
			if err = store.GetItems(ctx, ch); err != nil {
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
			bw := bytes.NewBuffer(nil)
			w.Header().Set("Content-Type", "text/plain")
			for pair := range ch {
				id := pair.int
				item := pair.Item

				fmt.Fprintf(bw, "%-12s %d\n", "ID", id)
				fmt.Fprintf(bw, "%-12s %s\n", "Description", item.Description)
				if item.Done {
					fmt.Fprintf(bw, "%-12s true\n", "Done")
				} else {
					fmt.Fprintf(bw, "%-12s false\n", "Done")
				}
				fmt.Fprintf(bw, "%-12s %s\n\n", "Time", item.Time.Format(time.RFC3339))
			}
			if err != nil {
				http.Error(w, "error fetching items"+err.Error(), http.StatusInternalServerError)
				return
			} else {
				io.Copy(w, bw)
			}
		default:
			http.Error(w, "unsupported type "+accepted, http.StatusUnsupportedMediaType)
			return
		}
	case "POST":
		_, err := store.database.ExecContext(ctx, "INSERT INTO items (description, time, done) VALUES (?, ?, ?)", item.Description, item.Time.Unix(), item.Done)
		if err != nil {
			http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	case "PATCH":
		_, err := store.database.ExecContext(ctx, "UPDATE items SET description = ?, done = ? WHERE id = ?", item.Description, item.Done, id)
		if err != nil {
			http.Error(w, "store.database error: "+err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	case "DELETE":
		_, err := store.database.ExecContext(ctx, "DELETE FROM items WHERE items.id = ?", id)
		if err != nil {
			http.Error(w, "store.database error: "+err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Store) GetItems(ctx context.Context, ch chan struct {
	int
	Item
}) (err error) {
	var conn *sql.Conn
	var rows *sql.Rows
	var t int64
	var e struct {
		int
		Item
	}

	if conn, err = s.database.Conn(ctx); err != nil {
		return
	}

	rows, err = conn.QueryContext(ctx, "SELECT id, description, time, done FROM items")
	if err != nil {
		return
	}

	for ok := true; ok && rows.Next(); {
		err = rows.Scan(&e.int, &e.Item.Description, &t, &e.Item.Done)
		if err != nil {
			return err
		}
		e.Item.Time = time.Unix(t, 0)
		select {
		case <-ctx.Done():
			ok = false
		case ch <- e:
		}
	}

	if err = rows.Close(); err != nil {
		conn.Close()
	} else {
		err = conn.Close()
	}

	return
}
