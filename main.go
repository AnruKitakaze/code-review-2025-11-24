package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var users = map[int]*User{}
var requestCount int
var mu sync.Mutex

func main() {
	log.Println("starting http server...")

	// фоновой логгер статистики
	go logStats(context.Background())

	http.HandleFunc("/user/create", handleCreateUser)
	http.HandleFunc("/user/list", handleListUsers)
	http.HandleFunc("/health", handleHealth)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("server stopped with error:", err.Error())
	}
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)

	incrementRequestCount()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "only POST allowed")
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "name is required")
		return
	}

	go func() {
		select {
		case <-ctx.Done():
			log.Println("create user timeout:", ctx.Err())
			return
		default:
			// имитируем какую-то работу
			time.Sleep(500 * time.Millisecond)
			id := len(users) + 1
			users[id] = &User{
				ID:   id,
				Name: name,
			}
			log.Println("created user", id, name)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "user creation started")
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	incrementRequestCount()

	mu.Lock()
	defer mu.Unlock()

	var result []*User
	for _, u := range users {
		result = append(result, u)
	}

	b, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to marshal")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func incrementRequestCount() {
	requestCount++
}

func logStats(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("logStats stopped:", ctx.Err())
			return
		default:
			time.Sleep(5 * time.Second)
			log.Println("stats: users =", len(users), "requests =", requestCount)
		}
	}
}
