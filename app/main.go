package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Setup struct {
	ID     string `json:"Id"`
	Secret string `json:"Secret"`
}

func main() {
	stat, _ := os.Stdin.Stat()

	// pass := flag.Bool("pass", false, "Pass-though stdin")

	// Timeout flag for process,  require FE encryption flag, TTL for stream (0 should delete right afterwards -> 24 hours)

	flag.Parse()

	if (stat.Mode() & os.ModeCharDevice) == 0 {

		reader := bufio.NewReader(os.Stdin)

		cli := &http.Client{Timeout: time.Hour * 24}
		req, err := http.NewRequest("GET", "http://127.0.0.1:3001/start", nil)

		res, err := cli.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		body, err := ioutil.ReadAll(res.Body)

		var data Setup
		json.Unmarshal(body, &data)

		log.Printf("ID: http://127.0.0.1:3000/%s | SECRET: %s", data.ID, data.Secret)
		req, err = http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:3001/push/%s", data.ID), reader)
		req.Header.Add("X-Stream-Secret", data.Secret)
		cli.Do(req)

		if err != nil {
			log.Fatal(err)
		}
	}
}
