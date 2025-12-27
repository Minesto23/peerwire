package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Minesto23/peerwire/internal/engine"
	"github.com/Minesto23/peerwire/internal/torrent"
)

//go:embed web/*
var webFS embed.FS

type Status struct {
	Running bool
	Percent float64
	Message string
}

var currentStatus = Status{Message: "Waiting for torrent..."}
var client *engine.Client

func main() {
	// Serve static files from embedded FS
	root, _ := fs.Sub(webFS, "web")
	fs := http.FileServer(http.FS(root))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			handleIndex(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	})

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/status", handleStatus)
	http.HandleFunc("/browse", handleBrowse)

	fmt.Println("Starting GUI at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleBrowse(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			path = "."
		} else {
			path = home
		}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type Folder struct {
		Name string
		Path string
	}

	var folders []Folder
	// Add parent directory option if not at root
	parent := filepath.Dir(path)
	if parent != path {
		folders = append(folders, Folder{Name: "..", Path: parent})
	}

	for _, e := range entries {
		if e.IsDir() && e.Name()[0] != '.' { // Skip hidden dirs
			folders = append(folders, Folder{
				Name: e.Name(),
				Path: filepath.Join(path, e.Name()),
			})
		}
	}

	resp := map[string]interface{}{
		"current": path,
		"folders": folders,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(webFS, "web/index.html")
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), 500)
		return
	}
	tmpl.Execute(w, nil)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 303)
		return
	}

	// Parsing multipart form (10MB max)
	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("torrent")
	if err != nil {
		currentStatus.Message = "Error: " + err.Error()
		http.Redirect(w, r, "/", 303)
		return
	}
	defer file.Close()

	destPath := r.FormValue("destination")
	if destPath == "" {
		destPath = "."
	}

	// Validate destination
	if info, err := os.Stat(destPath); err != nil || !info.IsDir() {
		currentStatus.Message = "Invalid Destination: " + destPath
		http.Redirect(w, r, "/", 303)
		return
	}

	// Save to temp
	tmpPath := filepath.Join(os.TempDir(), "upload.torrent")
	out, _ := os.Create(tmpPath)
	io.Copy(out, file)
	out.Close()

	// Parse
	f, _ := os.Open(tmpPath)
	spec, err := torrent.Parse(f)
	f.Close()

	if err != nil {
		currentStatus.Message = "Invalid Torrent: " + err.Error()
		http.Redirect(w, r, "/", 303)
		return
	}

	// Start Engine
	currentStatus.Running = true
	currentStatus.Percent = 0
	currentStatus.Message = "Initialize: " + spec.Info.Name

	fullOutputPath := filepath.Join(destPath, spec.Info.Name)

	params := engine.ClientParams{OutputPath: fullOutputPath}
	c, err := engine.NewClient(spec, params)
	if err != nil {
		currentStatus.Message = "Engine Error: " + err.Error()
		http.Redirect(w, r, "/", 303)
		return
	}
	client = c
	go func() {
		err := client.Download(func(done, total int) {
			currentStatus.Percent = float64(done) / float64(total) * 100
			currentStatus.Message = fmt.Sprintf("Downloading to %s... %.1f%%", destPath, currentStatus.Percent)
		})
		if err != nil {
			currentStatus.Message = "Failed: " + err.Error()
		} else {
			currentStatus.Message = "Download Complete!"
			currentStatus.Percent = 100
		}
	}()

	http.Redirect(w, r, "/", 303)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStatus)
}
