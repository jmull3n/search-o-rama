package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var pagesRepository Pages
var termsRepository Terms

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
// (static file server borrowed from gorilla mux readme examples)
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		return
	}
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

// This method is long, but until Product signs off on the
// functionality, I'd rather leave it all in a logical place
// instead of prematurely abstracting
func crawlPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Crawl Request recieved")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")

	if r.Method == http.MethodOptions {
		return
	}
	start := time.Now()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var input CrawlerInput
	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO: move these items to a config file or CLI param

	// how deep should we traverse the link graph
	depth := 3

	// how many crawls to run concurrently
	// We'll take it gently on our target servers
	concurrency := 15
	var crawlErrors []string
	log.Debug("Starting crawl...")
	crawledPages, termCount, errs := CrawlAndCatalog(input, concurrency, depth, &pagesRepository, &termsRepository)
	if errs != nil {
		for _, e := range errs {
			crawlErrors = append(crawlErrors, e.Error())
		}
	}

	elapsed := time.Since(start).Seconds()

	log.Debug("Crawl Request responded")

	json.NewEncoder(w).Encode(CrawlReponse{
		DurationSeconds: elapsed,
		WordsIndexed:    termCount,
		PagesCrawled:    crawledPages,
		CrawlErrors:     crawlErrors,
	})
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Search Request recieved")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")

	if r.Method == http.MethodOptions {
		return
	}
	start := time.Now()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var input SearchRequest
	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	results := termsRepository.GetTermPages(input.Term)
	elapsed := time.Since(start).Seconds()

	response, err := json.Marshal(
		SearchResponse{
			DurationSeconds: elapsed,
			Results:         results,
		})
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Debug("Search Request responded")

	w.Write(response)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Resest Request recieved")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")

	if r.Method == http.MethodOptions {
		return
	}

	// Danger zone...since this just yanks the rug out of any other in-process crawls
	log.Debug("Resetting Pages Repository")
	pagesRepository.Reset()

	log.Debug("Resetting Terms Index")
	termsRepository.Reset()

	log.Debug("Resest Request responded")

	w.WriteHeader(http.StatusOK)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func main() {
	// setup logging
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	// setup routes and stuff
	log.Debug("setting up routes")
	router := mux.NewRouter()
	router.HandleFunc("/api/health", healthHandler)
	router.HandleFunc("/api/crawl", crawlPageHandler).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/search", searchHandler).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/reset", resetHandler).Methods(http.MethodDelete, http.MethodOptions)

	router.Use(mux.CORSMethodMiddleware(router))

	// setup static file server
	log.Debug("setting up static file server")
	spa := spaHandler{staticPath: "static", indexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:7250",
	}
	log.Debug("launching webserver!")
	log.Fatal(srv.ListenAndServe())
}
