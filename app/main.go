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

	flag.Parse()

	if (stat.Mode() & os.ModeCharDevice) == 0 {

		reader := bufio.NewReader(os.Stdin)

		cli := &http.Client{Timeout: 1 * time.Second}
		req, err := http.NewRequest("GET", "http://127.0.0.1:3000/start", nil)

		res, err := cli.Do(req)
		body, err := ioutil.ReadAll(res.Body)

		var data Setup
		json.Unmarshal(body, &data)
		

		// if (*pass) {
		// 	go io.Copy(os.Stdout, os.Stdin)
		// }
		// resp, err := http.Get("http://127.0.0.1:3000/start")
		log.Printf("ID: %s | SECRET: %s", data.ID, data.Secret)
		req, err = http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:3000/s/%s", data.ID), reader)
		req.Header.Add("X-Stream-Secret", data.Secret)
		cli.Do(req)
		defer req.Body.Close()

		// scanner := bufio.NewScanner(os.Stdin)

		// for scanner.Scan() {
		// 	if (*pass) {
		// 		fmt.Println(scanner.Text())
		// 	}

		// 	buf = append(buf, scanner.Bytes()...)
			
		// }

		if err != nil {
			log.Fatal(err)
		}
	}
}
