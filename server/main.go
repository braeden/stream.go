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

func startLogging(w http.ResponseWriter, r *http.Request) {
	// Start redis queue and setup appropriate subscriber?
	id := make([]byte, 5)
	secret := make([]byte, 32)
	rand.Read(id)
	rand.Read(secret)

	setup := Setup{
		ID:     hex.EncodeToString(id),
		Secret: hex.EncodeToString(secret),
	}

	err := client.Set(ctx, setup.ID, setup.Secret, time.Duration(time.Hour * 24)).Err()
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
	for scanner.Scan() {
		log.Printf("Got data: %s", scanner.Text())
		// push buf to redis
		err := client.XAdd(ctx, &redis.XAddArgs{
			Stream:       fmt.Sprintf("%s-stream", id),
			MaxLen:       0,
			MaxLenApprox: 0,
			ID:           "",
			Values: map[string]interface{}{
				"entry": string(scanner.Text()),
			},
		}).Err()

		if err != nil {
			w.WriteHeader(http.StatusTeapot)
			log.Fatal("Redis died: ", err)
			return
		}
	}
	log.Printf("Closed connection for %s", id)
	// client.Del(id)
	// Shut down and cleanup
}

func main() {

	client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", "127.0.0.1", "6379"),
	})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Unable to connect to Redis", err)
	}
	log.Println("Connected to Redis server")

	r := mux.NewRouter()
	r.HandleFunc("/start", startLogging).Methods("GET")
	r.HandleFunc("/s/{id}", addLogs).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:3000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Web server is coming up: %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

// We basically will have two endpoints here

// first we connect to redis (create a new stream) (get an ID and a generate a secret key)

// endpoint #1 /start (GET)

// endpoint #2 add (key, log) POST (we should look into how the app can open a streaming conneciton with us)

// endpoint #3, socket IO communication, https://github.com/googollee/go-socket.io (or in general manage the creation & deletion of rooms)
// create a subscribie pool, 1 redis subscriber -> updates to all rooms (maybe spwan this as a thread?)

// chat from client (URL (key), last line number, ID for user)
// in return we send back all the data it needs (maybe we want to do this via HTTP (seems easier than negotatiing 2-ways))
