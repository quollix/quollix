package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	version        = os.Getenv("VERSION")
	port           = os.Getenv("PORT")
	AuthCookieName = "quollix-auth"
)

func main() {
	http.HandleFunc("/", versionHandler)
	http.HandleFunc("/env/", envHandler)
	http.HandleFunc("/save-string", saveStringHandler)
	http.HandleFunc("/read-string", readStringHandler)
	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil)) // #nosec G114 (CWE-676): Use of net/http serve function that has no support for setting timeouts; sample app is only relevant for testing
}

func envHandler(w http.ResponseWriter, r *http.Request) {
	requestedEnvName := strings.TrimPrefix(r.URL.Path, "/env/")
	if requestedEnvName == "" {
		http.Error(w, "unknown env variable", http.StatusBadRequest)
		return
	}

	if _, err := w.Write([]byte(os.Getenv(requestedEnvName))); err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		log.Printf("Error writing response: %v", err)
	}
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(AuthCookieName) != "" {
		message := fmt.Sprintf("error, the '%s' header should never be set", AuthCookieName)
		http.Error(w, message, http.StatusBadRequest)
		return
	}
	if len(r.Cookies()) > 0 {
		http.Error(w, "error, the request should not contain any cookies", http.StatusBadRequest)
		return
	}
	if len(r.URL.Query()) > 0 {
		http.Error(w, "error, the request should not contain any query parameters", http.StatusBadRequest)
		return
	}

	_, err := w.Write([]byte("this is version " + version))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

func saveStringHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusBadRequest)
		log.Printf("Error reading request body: %v", err)
		return
	}
	err = os.WriteFile("/data/string.txt", data, 0600)
	if err != nil {
		http.Error(w, "error saving string to file", http.StatusInternalServerError)
		log.Printf("Error writing string to file: %v", err)
		return
	}
}

func readStringHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("/data/string.txt")
	if err != nil {
		log.Printf("Error reading string file: %v", err)
		http.Error(w, "error reading string file", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "error writing response", http.StatusInternalServerError)
		return
	}
}
