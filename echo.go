package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var exitCode = 200

func curl(url string) (string, error) {
	cmd := exec.Command("curl", "--http0.9", "-s", url)
	res, err := cmd.CombinedOutput()
	return string(res), err
}

func main() {
	// CRASH: HH:MM else just crash
	if date := os.Getenv("CRASH"); date != "" { // Just crash!
		// 58859 20-01-11 21:28:24 00 0 0 129.3 UTC(NIST) *
		// get time: curl http://time.nist.gov:13
		if len(date) > 3 && date[2] == ':' {
			if now, err := curl("http://time.nist.gov:13"); err == nil {
				parts := strings.SplitN(now, " ", 4)
				if len(parts) > 3 {
					now = parts[2] // Just time part
					now = now[:len(date)]
					if now > date {
						fmt.Printf("Timed crashing on demand\n")
						os.Exit(1)
					}
				}
			} else {
				fmt.Printf("Curl: %s\n%s\n", now, err)
			}
		} else {
			fmt.Printf("Crashing on demand\n")
			os.Exit(1)
		}
	}

	// Just so we can better control when we use the stream
	cronCount := int64(0)
	echoCount := int64(0)

	hostname := os.Getenv("HOSTNAME")
	msg := os.Getenv("MSG")
	if msg == "" {
		msg = "Hi from echo"
	}
	rev := ""
	if tmp := os.Getenv("K_REVISION"); tmp != "" {
		rev = " rev: " + tmp
	}

	envs := os.Environ()
	sort.StringSlice(envs).Sort()
	env := strings.Join(envs, "\n")
	fmt.Printf("Envs:\n%s\n", env)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var body []byte

		if r.Body != nil {
			body, _ = ioutil.ReadAll(r.Body)
		}

		fmt.Printf("%s:\n%s %s\n", time.Now().String(), r.Method, r.URL)
		headers := []string{}
		for k, _ := range r.Header {
			headers = append(headers, k)
		}
		sort.StringSlice(headers).Sort()
		for _, k := range headers {
			fmt.Printf("%s: %v\n", k, r.Header[k])
		}
		fmt.Printf("Body:\n%s\n\n", string(body))

		if r.URL.Path == "/stats" {
			fmt.Fprintf(w, "%d/%d\n", echoCount, cronCount)
		} else {
			// ?curl=host
			curlAddr := r.URL.Query().Get("curl")
			if curlAddr != "" {
				output, err := curl(curlAddr)
				output = strings.TrimSpace(output)
				if err != nil {
					w.WriteHeader(500)
					fmt.Fprintf(w, "Error curling(%s): %s - %s\n", curlAddr,
						err, output)
					return
				}
				fmt.Fprintf(w, "Curl: %s\n", output)
			}

			sleep := r.URL.Query().Get("sleep")
			if sleep == "" {
				sleep = os.Getenv("SLEEP")
			}

			if sleep != "" {
				len, _ := strconv.Atoi(sleep)
				fmt.Printf("Sleeping %d\n", len)
				time.Sleep(time.Duration(len) * time.Second)
			}
			if r.URL.Query().Get("crash") != "" {
				fmt.Printf("Crashing...\n")
				os.Exit(1)
			}

			ec := exitCode
			if t := r.URL.Query().Get("exit"); t != "" {
				if s, err := strconv.Atoi(t); err == nil {
					ec = s
				}
			}
			w.WriteHeader(ec)
			fmt.Printf("Exit(%d)\n", ec)

			if len(body) == 0 {
				fmt.Fprintf(w, "%s (host: %s%s)\n", msg, hostname, rev)
			} else {
				fmt.Fprintf(w, string(body)+"\n")
			}

			if strings.Contains(string(body), "cron") {
				atomic.AddInt64(&cronCount, 1)
			} else {
				atomic.AddInt64(&echoCount, 1)
			}

		}
	})

	// HTTP_DELAY will pause for 'delay' seconds before starting the
	// HTTP server. This is useful for simulating a long readiness probe
	if delay := os.Getenv("HTTP_DELAY"); delay != "" {
		sec, _ := strconv.Atoi(delay)
		if sec != 0 {
			fmt.Printf("Sleeping %d seconds before starting server...\n", sec)
			time.Sleep(time.Duration(sec) * time.Second)
		}
	}

	if exit := os.Getenv("EXIT"); exit != "" {
		if s, err := strconv.Atoi(exit); err == nil {
			exitCode = s
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Listening on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
