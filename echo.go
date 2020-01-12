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
	"time"
)

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
						os.Exit(1)
					}
				}
			} else {
				fmt.Printf("Curl: %s\n%s\n", now, err)
			}
		} else {
			os.Exit(1)
		}
	}

	cronCount := 0
	echoCount := 0

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

		fmt.Printf("%s:\n%s %s\nHeaders:\n%s\n\nBody:\n%s\n\n",
			time.Now().String(), r.Method, r.URL, r.Header, string(body))

		if r.URL.Path == "/stats" {
			fmt.Fprintf(w, "%d/%d\n", echoCount, cronCount)
		} else {
			if len(body) == 0 {
				if t := r.URL.Query().Get("sleep"); t != "" {
					len, _ := strconv.Atoi(t)
					time.Sleep(time.Duration(len) * time.Second)
				}
				fmt.Fprintf(w, "%s (host: %s%s)\n", msg, hostname, rev)
			} else {
				fmt.Fprintf(w, string(body)+"\n")
			}

			if strings.Contains(string(body), "cron") {
				cronCount++
			} else {
				echoCount++
			}

		}
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
