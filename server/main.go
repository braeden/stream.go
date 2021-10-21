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
	// socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	// "github.com/rs/cors"
)

var client *redis.Client
// var socketSvr *socketio.Server

var ctx = context.Background()

type Setup struct {
	ID     string `json:"Id"`
	Secret string `json:"Secret"`
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

	// streamName := fmt.Sprintf("%s-stream", id)
	// terminate := make(chan bool)
	// go func() {
	// 	redisMessageID := "0"
	// 	for {
	// 		select {
	// 		case <-terminate:
	// 			log.Println("GO ROUTIne stopped")
	// 			return
	// 		default:
	// 			res, err := client.XRead(ctx, &redis.XReadArgs{
	// 				Streams: []string{streamName, redisMessageID},
	// 				Block:   1 * time.Second,
	// 			}).Result()
	// 			if err != nil {
	// 				// Didn't get data in time (let's check term status)
	// 				continue
	// 			}
	// 			for _, e := range res {
	// 				for _, m := range e.Messages {
	// 					log.Println("Go Routine msg")

	// 					redisMessageID = m.ID
	// 					// We only want this is we're not using redis adapter I think?
	// 					// if server.RoomLen("/", id) > 0 {
	// 					j, _ := json.Marshal(m.Values)
	// 					socketSvr.BroadcastToRoom("/", id, "log", string(j))
	// 					// }
	// 				}
	// 			}
	// 		}
	// 	}
	// }()

	i := 0
	for scanner.Scan() {
		log.Printf("Got data: %s", scanner.Text())
		err := client.XAdd(ctx, &redis.XAddArgs{
			Stream:       fmt.Sprintf("%s-stream", id),
			MaxLen:       0,
			MaxLenApprox: 0,
			ID:           "",
			Values: map[string]interface{}{
				"entry":   scanner.Text(),
				"lineNum": i,
			},
		}).Err()

		if err != nil {
			w.WriteHeader(http.StatusTeapot)
			return
		}
		err = client.Publish(ctx, fmt.Sprintf("%s-pubsub", id), scanner.Text()).Err()

		if err != nil {
			w.WriteHeader(http.StatusTeapot)
			return
		}
		i++
	}
	// client.Del(ctx, fmt.Sprintf("%s-stream", id))
	// terminate <- true
	client.Del(ctx, id)
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

	// socketSvr = socketio.NewServer(nil)

	// _, err = socketSvr.Adapter(&socketio.RedisAdapterOptions{
	// 	Addr:   addr,
	// 	Prefix: "socket.io",
	// })

	if err != nil {
		log.Fatal("Unable to connect to Redis for socket.io", err)
	}

	// socketSvr.OnConnect("/", func(s socketio.Conn) error {
	// 	// s.SetContext("")
	// 	fmt.Println("connected:", s.ID())
	// 	return nil
	// })

	// socketSvr.OnDisconnect("/", func(s socketio.Conn, reason string) {
	// 	fmt.Println("closed", reason)
	// })
	// go socketSvr.Serve()
	// defer socketSvr.Close()

	r := mux.NewRouter()
	r.HandleFunc("/start", startLogging).Methods("GET")
	r.HandleFunc("/push/{id}", addLogs).Methods("POST")
	// r.Handle("/socket.io/", socketSvr)

	// c := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"http://localhost:3000"},
	// 	AllowCredentials: true,
	// })

	// handler := c.Handler(r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:3001",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  24 * time.Hour,
	}

	log.Printf("Web server is coming up: %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
