package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Config struct Model
type Config struct {
	URL      string        `yaml:"url"`
	Method   string        `yaml:"method"`
	Requests int           `yaml:"requests"`
	Workers  int           `yaml:"workers"`
	Timeout  time.Duration `yaml:"timeout"`
	body     string        `yaml:"body"`
	header   string        `yaml:"header"`
}

var config *Config
var starting = time.Now()

// Custom user agent.
const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/53.0.2785.143 " +
		"Safari/537.36"
)

func worker(id int, jobs <-chan int) {
	for j := range jobs {

		start := time.Now()

		client := &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				// Dial:            (&net.Dialer{Timeout: 0, KeepAlive: 0}).Dial,
				// Dial: func(netw, addr string) (net.Conn, error) {
				// 	deadline := time.Now().Add(25 * time.Second)
				// 	c, err := net.DialTimeout(netw, addr, time.Second*20)
				// 	if err != nil {
				// 		return nil, err
				// 	}
				// 	c.SetDeadline(deadline)
				// 	return c, nil
				// },
				TLSHandshakeTimeout: config.Timeout * time.Second,
			},
			Timeout: config.Timeout * time.Second,
		}

		req, _ := http.NewRequest(strings.ToUpper(config.Method), config.URL, nil)
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Connection", "close")

		resp, err := client.Do(req)

		if err != nil {

			log.Println("\x1b[33;1m|", err, "|\x1b[37;1m", strings.ToUpper(config.Method), time.Since(start), "| Worker", id, "| Job", j)
		} else {

			log.Println("\x1b[32m|", resp.Status, "|\x1b[37;1m", strings.ToUpper(config.Method), time.Since(start), "| Worker", id, "| Job", j)
			resp.Body.Close()
		}

	}
}

func main() {

	// CLEAR THE TERM
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()

	// HANDLE INTERRUPT
	HandleInterrupt()

	fmt.Println(`
    _  _     ____  _____  _  ____    ____  _____ ____  _     _____ ____  _____  _____ ____ 
   / |/ \ /\/ ___\/__ __\/ \/  __\  /  __\/  __//  _ \/ \ /\/  __// ___\/__ __\/  __//  __\
   | || | |||    \  / \  | || | //  |  \/||  \  | / \|| | |||  \  |    \  / \  |  \  |  \/|
/\_| || \_/|\___ |  | |  | || |_\\  |    /|  /_ | \_\|| \_/||  /_ \___ |  | |  |  /_ |    /
\____/\____/\____/  \_/  \_/\____/  \_/\_\\____\\____\\____/\____\\____/  \_/  \____\\_/\_\`)

	fmt.Println("\n\x1b[33;1mTrying to Unmarshal the config file 🤪 \x1b[37;1m")

	source, err := ioutil.ReadFile("config.yaml")
	ErrorPanic(err, "Config not Found.")

	err = yaml.Unmarshal(source, &config)
	ErrorPanic(err, "Can't Unmarshal the config file.")

	switch strings.ToLower(config.Method) {

	case "get":

	case "post":

	case "put":

	default:
		log.Fatal("\n\x1b[33;1\nmMethod not Supported 😭 \x1b[37;1m")
	}

	fmt.Println("\n\x1b[33;1mMethod selected :\x1b[37;1m", config.Method)

	if config.URL == "" {

		log.Fatal("\n\x1b[33;1m\nYou need to specified an url 😭  \x1b[37;1m")
	}

	fmt.Println("\n\x1b[33;1mTarget url :\x1b[37;1m", config.URL)

	if config.Requests == 0 {
		fmt.Println("\n\x1b[33;1mRequests not defined, using default value : 50000 🤪 \x1b[37;1m")
		config.Requests = 50000
	}

	fmt.Println("\n\x1b[33;1mRequest limit : \x1b[37;1m", config.Requests)

	if config.Workers == 0 {
		fmt.Println("\n\x1b[33;1mWorkers pool not defined, using default value : 100 🤪 \x1b[37;1m")
		config.Workers = 100
	}

	fmt.Println("\n\x1b[33;1mWorkers pool limit : \x1b[37;1m", config.Workers)

	fmt.Println("\n\x1b[33;1mTimeout selected : \x1b[37;1m", config.Timeout*time.Second, "\n")

	jobs := make(chan int, 100)

	// ICI ON SPAWN LES WORKERS
	for w := 1; w <= config.Workers; w++ {
		go worker(w, jobs)
	}

	// ON AJOUTE DES TASKS A LA QUEUE
	for j := 0; j <= config.Requests; j++ {
		jobs <- j
	}
}

func HandleInterrupt() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-c
		Shutdown(sig)
	}()
}

func Shutdown(sig os.Signal) {

	log.Println("Captured :", sig)
	var timeSinceBegining = time.Since(starting)
	log.Println("Time since the begining :", timeSinceBegining)
	os.Exit(1)
}

func ErrorPanic(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
