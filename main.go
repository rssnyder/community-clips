package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

var (
	psql string
)

func init() {
	flag.StringVar(&psql, "psql", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable", "postgres connection string")
	flag.Parse()
}

// Clips stores important stuff
type Clips struct {
	database *sql.DB
}

func main() {

	// use postgres for job queue
	db, err := sql.Open("postgres", psql)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("database: connected")

	clips := Clips{
		database: db,
	}

	// use http for router
	mux := http.NewServeMux()

	mux.HandleFunc("/", clips.index)
	mux.HandleFunc("/clips", clips.clips)
	mux.HandleFunc("/clips/new", clips.clipsNew)

	fmt.Println("api: serving")
	err = http.ListenAndServe(":3000", mux)
	log.Fatal(err)
}

func (c *Clips) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodGet {

		fmt.Println("someone is at the door!")
		w.WriteHeader(http.StatusAccepted)
	} else {
		defaultReply(w)
	}
}

// clips returns all clips in the database
func (c *Clips) clips(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/clips" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodGet {

		jobs, err := c.Jobs()
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(jobs)
	} else {
		defaultReply(w)
	}
}

// clipsNew submits a new clip to the database for processing
func (c *Clips) clipsNew(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/clips/new" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "unable to parse data", http.StatusBadRequest)
			return
		}

		sourceLink := r.Form.Get("source_link")
		if sourceLink == "" {
			http.Error(w, "source_link required", http.StatusBadRequest)
			return
		}
		start := r.Form.Get("start")
		if start == "" {
			http.Error(w, "start required", http.StatusBadRequest)
			return
		}
		duration := r.Form.Get("duration")
		if duration == "" {
			http.Error(w, "duration required", http.StatusBadRequest)
			return
		}

		startInt, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			http.Error(w, "start should be an integer", http.StatusBadRequest)
			return
		}

		durationInt, err := strconv.ParseInt(duration, 10, 64)
		if err != nil {
			http.Error(w, "duration should be an integer", http.StatusBadRequest)
			return
		}

		// check the exact clip dosnt exist
		exists, err := c.FindJob(sourceLink, startInt, durationInt)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if exists {
			w.WriteHeader(http.StatusAlreadyReported)
			return
		}

		// create a new job and add to the queue
		job, err := c.InsertJob(sourceLink, startInt, durationInt)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	} else {
		defaultReply(w)

	}
}

// defaultReply tells them whats up
func defaultReply(w http.ResponseWriter) {
	w.Header().Set("Allow", "POST, OPTIONS")
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
