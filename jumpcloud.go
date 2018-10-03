package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/scruffycatherder/jumpcloud/cache"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

var port = flag.Int("port", 8080, "The port number that the HTTP server will listen on.")
var gracefulShutdown = make(chan os.Signal)

// Used to count # of calls to /hash AND to generate indexes into the cache.
// hashCount must only be accessed via atomic functions from sync/atomic.
var hashCount uint64 = 0

// Used to track rolling average of execution time spent calculating hash values.
var hashAvg uint64 = 0
var hashTotal uint64 = 0

var hashes *cache.HashCache = cache.NewHashCache()

// Regular expression matches "/hash", "/hash/" and "/hash/{key}"
// Using path.FindStringSubMatch on a URL will return {key} if it exists.
var path = regexp.MustCompile("^/hash/*([0-9]+)*$")

// Handler for requests to generate a hash of a password.
func hashHandler(w http.ResponseWriter, r *http.Request) {
	match := path.FindStringSubmatch(r.URL.Path)
	var result string = ""
	if match != nil {
		// our regular expression match ensures that this parse never fails
		key, _ := strconv.ParseUint(match[1], 10, 64)
		hash, ok := hashes.Get(key)
		// TODO(bobl): error handling
		if ok {
			result = hash
		}
	}
	_, err := fmt.Fprintln(w, result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for requests for execution statistics.
func statsHandler(w http.ResponseWriter, r *http.Request) {
	count := atomic.LoadUint64(&hashCount)
	totalTime := atomic.LoadUint64(&hashTotal)

	// NOTE:  The calculation of average here is an approximation at best.
	// Although we maintain atomic access to hashCount and hashTotal individually, we
	// do not actually read and write them atomically together.  Given the 5 second
	// sleep time before calculating hashes, the values will be out of sync.  Results
	// will be skewed significantly early after server startup, but will settle to
	// a reasonably accurate stat over time.
	// We could make enforce atomic access to these metrics as well, but it would
	// likely incur a performance hit, and seemed a bit excessive for this exercise.

	result := fmt.Sprintf("{\"total\": %d, \"average\": %d}", count, totalTime/count)

	_, err := fmt.Fprintln(w, result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func hashGeneratorHandler(w http.ResponseWriter, r *http.Request) {
	pass := r.FormValue("password")
	count := atomic.AddUint64(&hashCount, 1)

	result := strconv.FormatUint(count, 10)
	// GoRoutine to generate hash after 5 seconds and cache it.
	go func() {

		// Wait 5 seconds before performing the hash
		// TODO(bobl): make this a flag
		time.Sleep(5 * time.Second)

		// TODO(bobl): Clarify the requirement about average processing time metrics.
		// They mentioned entire time processing the hash endpoint.  Taken literally,
		// that probably means the duration from socket connect to release on any call
		// to /hash OR /hash/{num}.
		// However, it strikes me that that more interesting metric (at least initially)
		// is the time spent doing the heavy calculations, so that's what I'll start
		// with here.
		// Note that I'm explicitly excluding the sleep time in the instrumentation here.
		// My thought here is that the sleep time will vastly dwarf the actual execution
		// time, making the metric less useful if the sleep time is included.  Regardless,
		// this can easily be changed (and new metrics added) as required.
		start := time.Now()

		checksum := sha512.Sum512([]byte(pass))
		hashed := base64.StdEncoding.EncodeToString(checksum[:])
		hashes.Put(count, hashed)
		var elapsed uint64 = uint64(time.Since(start).Nanoseconds())

		atomic.AddUint64(&hashTotal, elapsed)
	}()

	_, err := fmt.Fprintln(w, result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	log.Println("JumpCloud Go exercise Starting...")
	flag.Parse()
}

func main() {
	log.Println("HTTP Server Starting on port: ", *port)

	signal.Notify(gracefulShutdown, syscall.SIGTERM)
	signal.Notify(gracefulShutdown, syscall.SIGINT)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", *port),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Println("Signal Handler starting")
		sig := <-gracefulShutdown
		log.Println("SignalHandler caught signal: ", sig)
		log.Println("Wait for 2 seconds to finish processing.")
		server.Shutdown(context.Background())
		log.Println("Server shutdown complete")
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	http.HandleFunc("/hash/", hashHandler)
	http.HandleFunc("/hash", hashGeneratorHandler)
	http.HandleFunc("/stats/", statsHandler)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		//Error starting or closing listener:
		log.Println("HTTP server ListenAndServe: %v", err)
	}
}
