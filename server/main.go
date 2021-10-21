package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var client *redis.Client

var ctx = context.Background()

type Setup struct {
	ID     string `json:"Id"`
	Secret string `json:"Secret"`
}

type Entry struct {
	Text string `json:"text"`
	Line int    `json:"line"`
	Id   string `json:"id"`
}

func startLogging(w http.ResponseWriter, r *http.Request) {
	// Start redis queue and setup appropriate subscriber?
	id := make([]byte, 4)
	secret := make([]byte, 32)
	rand.Read(id)
	rand.Read(secret)

	setup := Setup{
		ID:     hex.EncodeToString(id),
		Secret: hex.EncodeToString(secret),
	}

	err := client.Set(ctx, setup.ID, setup.Secret, time.Duration(time.Hour*24)).Err()
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		return
	}
	json.NewEncoder(w).Encode(setup)
}

func addLogs(w http.ResponseWriter, r *http.Request) {
	scanner := bufio.NewScanner(r.Body)
	id := mux.Vars(r)["id"]
	secret := r.Header.Get("X-Stream-Secret")
	val, err := client.Get(ctx, id).Result()
	if err != nil || val != secret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	i := 0
	for scanner.Scan() {
		log.Printf("Got data: %s", scanner.Text())

		streamKey, err := client.XAdd(ctx, &redis.XAddArgs{
			Stream:       fmt.Sprintf("%s-stream", id),
			MaxLen:       0,
			MaxLenApprox: 0,
			ID:           "",
			Values: map[string]interface{}{
				"text": scanner.Text(),
				"line": i,
			},
		}).Result()
		// We use pub/sub in combination w/ streams to avoiding blocking
		// on XREAD -> we coorelate them using streamKey (to request past ranges)
		entry := Entry{
			Line: i,
			Text: scanner.Text(),
			Id:   streamKey,
		}
		// This is a different ID (this is the
		// stream key for this particular message instead of the UID)
		if err != nil {
			w.WriteHeader(http.StatusTeapot)
			return
		}
		bytes, err := json.Marshal(entry)
		err = client.Publish(ctx, fmt.Sprintf("%s-pubsub", id), string(bytes)).Err()

		if err != nil {
			w.WriteHeader(http.StatusTeapot)
			return
		}
		i++
	}
	log.Printf("Closed connection for %s", id)
}

func main() {

	addr := fmt.Sprintf("%s:%s", "127.0.0.1", "6379")
	client = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Unable to connect to Redis", err)
	}
	log.Println("Connected to Redis server")
	if err != nil {
		log.Fatal("Unable to connect to Redis for socket.io", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/start", startLogging).Methods("GET")
	r.HandleFunc("/push/{id}", addLogs).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:3001",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  24 * time.Hour,
	}

	log.Printf("Web server is coming up: %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
