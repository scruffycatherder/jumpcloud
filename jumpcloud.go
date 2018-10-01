// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
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

var hashes *cache.HashCache = cache.NewHashCache()

// Regular expression matches "/hash", "/hash/" and "/hash/{key}"
// Using path.FindStringSubMatch on a URL will return {key} if it exists.
var path = regexp.MustCompile("^/hash/*([0-9]+)*$")

func hashHandler(w http.ResponseWriter, r *http.Request) {
	match := path.FindStringSubmatch(r.URL.Path)
	var result string = ""
	if match != nil {
		log.Println("match length: ", len(match))
		log.Println("Retreiving hash")
		//log.Println(match)
		// our regular expression match ensures that this parse never fails
		key, _ := strconv.ParseUint(match[1], 10, 64)
		hash, err := hashes.Get(key)
		// TODO(bobl): error handling
		if err {
			log.Println("error: ", err)
		}
		result = hash
	}
	log.Println(result)
	_, err := fmt.Fprintln(w, result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func hashGeneratorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Generating hash")
	pass := r.FormValue("password")
	// TODO get this from atomic global
	count := atomic.AddUint64(&hashCount, 1)

	result := strconv.FormatUint(count, 10)
	// GoRoutine to generate hash after 5 seconds and cache it.
	go func() {
		// Wait 5 seconds before performing the hash
		// TODO(bobl): make this a flag
		time.Sleep(5 * time.Second)
		// TODO(bobl): Clarify exactly which SHA512 variant they wanted.
		// TODO(bobl): Add a salt?
		sha_512 := sha512.New512_224()
		hashed := base64.StdEncoding.EncodeToString(sha_512.Sum([]byte(pass)))
		hashes.Put(count, hashed)
		log.Println(hashed)
	}()

	log.Println(result)
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
		log.Println("Wait for 2 second to finish processing.")
		server.Shutdown(context.Background())
		log.Println("Server shutdown complete")
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	http.HandleFunc("/hash/", hashHandler)
	http.HandleFunc("/hash", hashGeneratorHandler)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		//Error starting or closing listener:
		log.Println("HTTP server ListenAndServe: %v", err)
	}
}
