package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	k8s "k8s.io/client-go/kubernetes"
)

const DEFAULT_PORT = ":8080"

// TODO: Remove once `alpha` cease to exist
const DEFAULT_UNIDLE_KEY_LABEL = "host"

var (
	logger         *log.Logger
	k8sClient      k8s.Interface
	indexTemplates *template.Template
	err            error
	UnidleKeyLabel string
)

func init() {
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	// parse HTML template
	indexTemplates, err = template.New("").ParseFiles(
		"templates/content.html",
		"templates/javascript.js",
		"templates/throbber.html",
		"templates/layout.html",
	)
	if err != nil {
		logger.Fatalf("Error parsing template: %s", err)
	}
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		logger.Printf("$PORT not set. Defaulting to '%s'", DEFAULT_PORT)
		port = DEFAULT_PORT
	}
	home, ok := os.LookupEnv("HOME")
	if !ok {
		logger.Fatalf("$HOME not set. It couldn't determine HOME directory.")
	}

	// NOTE: Default to `host` for retro-compatibility with `alpha` cluster
	// TODO: Remove logic and always use `unidle-key` label once migration to
	//       `prod`/new domain is completed
	UnidleKeyLabel, ok = os.LookupEnv("UNIDLE_KEY_LABEL")
	if !ok {
		logger.Printf("$UNIDLE_KEY_LABEL not set. Defaulting to '%s'", DEFAULT_UNIDLE_KEY_LABEL)
		UnidleKeyLabel = DEFAULT_UNIDLE_KEY_LABEL
	}

	k8sClient, err = KubernetesClient(filepath.Join(home, ".kube", "config"))
	if err != nil {
		log.Fatalf("Failed to create k8s client: %s", err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/events/", eventsHandler)
	http.HandleFunc("/healthz", healthzHandler)

	logger.Printf("Starting server on port %s...", port)
	server := &http.Server{
		Addr:         port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 2 * time.Minute,
		IdleTimeout:  2 * time.Minute,
	}
	log.Fatal(server.ListenAndServe())
}
