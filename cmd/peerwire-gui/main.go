package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Minesto23/peerwire/internal/engine"
	"github.com/Minesto23/peerwire/internal/torrent"
)

type Status struct {
	Running bool
	Percent float64
	Message string
}

var currentStatus = Status{Message: "Waiting for torrent..."}
var client *engine.Client

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/status", handleStatus)

	fmt.Println("Starting GUI at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	const tmplStr = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PeerWire Client</title>
    <style>
        :root {
            --bg-color: #0f0f13;
            --card-bg: #1a1a23;
            --text-primary: #ffffff;
            --text-secondary: #a0a0b0;
            --accent-color: #6c5ce7;
            --accent-gradient: linear-gradient(135deg, #6c5ce7 0%, #a29bfe 100%);
            --progress-bg: #2d2d3a;
            --error-color: #ff6b6b;
            --success-color: #00b894;
        }

        body {
            margin: 0;
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-primary);
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            background-image: radial-gradient(circle at 10% 20%, rgba(108, 92, 231, 0.1) 0%, transparent 20%);
        }

        .container {
            width: 100%;
            max-width: 500px;
            padding: 20px;
        }

        .card {
            background-color: var(--card-bg);
            border-radius: 16px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
            padding: 30px;
            text-align: center;
            border: 1px solid rgba(255, 255, 255, 0.05);
        }

        h1 {
            font-size: 24px;
            font-weight: 700;
            margin-bottom: 30px;
            background: var(--accent-gradient);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            letter-spacing: -0.5px;
        }

        .upload-area {
            border: 2px dashed rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            padding: 40px 20px;
            margin-bottom: 20px;
            transition: all 0.3s ease;
            position: relative;
            cursor: pointer;
        }

        .upload-area:hover {
            border-color: var(--accent-color);
            background: rgba(108, 92, 231, 0.05);
        }

        input[type=file] {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            opacity: 0;
            cursor: pointer;
        }

        .upload-label {
            color: var(--text-secondary);
            font-size: 14px;
            pointer-events: none;
        }

        .upload-icon {
            font-size: 32px;
            margin-bottom: 10px;
            display: block;
            color: var(--accent-color);
        }

        button {
            width: 100%;
            padding: 14px;
            border: none;
            border-radius: 8px;
            background: var(--accent-gradient);
            color: white;
            font-weight: 600;
            font-size: 16px;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(108, 92, 231, 0.4);
        }

        button:active {
            transform: translateY(0);
        }

        .status-panel {
            margin-top: 30px;
            text-align: left;
            background: rgba(0,0,0,0.2);
            padding: 15px;
            border-radius: 8px;
        }

        .status-text {
            color: var(--text-secondary);
            font-size: 13px;
            margin-bottom: 8px;
            display: flex;
            justify-content: space-between;
        }
        
        #msg {
            color: var(--text-primary);
            font-weight: 500;
        }

        .progress-track {
            height: 8px;
            background: var(--progress-bg);
            border-radius: 4px;
            overflow: hidden;
            margin-top: 5px;
        }

        .progress-bar {
            height: 100%;
            background: var(--accent-gradient);
            width: 0%;
            border-radius: 4px;
            transition: width 0.3s ease;
            box-shadow: 0 0 10px rgba(108, 92, 231, 0.5);
        }

        .percentage {
            float: right;
            color: var(--accent-color);
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <h1>PeerWire</h1>
            
            <form action="/upload" method="post" enctype="multipart/form-data" id="uploadForm">
                <div class="upload-area">
                    <input type="file" name="torrent" accept=".torrent" required id="fileInput" onchange="updateFileName()">
                    <span class="upload-icon">📂</span>
                    <span class="upload-label" id="fileLabel">Drop .torrent file or click to browse</span>
                </div>
                <button type="submit">Download</button>
            </form>
            
            <div class="status-panel">
                <div class="status-text">
                    <span id="msg">{{.Message}}</span>
                    <span class="percentage" id="percent">0%</span>
                </div>
                <div class="progress-track">
                    <div id="bar" class="progress-bar" style="width: {{.Percent}}%"></div>
                </div>
            </div>
        </div>
    </div>
    
    <script>
        function updateFileName() {
            const input = document.getElementById('fileInput');
            const label = document.getElementById('fileLabel');
            if (input.files && input.files.length > 0) {
                label.innerText = input.files[0].name;
            }
        }

        setInterval(() => {
            fetch('/status').then(r => r.json()).then(data => {
                document.getElementById('msg').innerText = data.Message;
                document.getElementById('percent').innerText = Math.round(data.Percent) + '%';
                document.getElementById('bar').style.width = data.Percent + '%';
                
                if (data.Percent >= 100) {
                     document.getElementById('bar').style.background = 'var(--success-color)';
                }
            });
        }, 800);
    </script>
</body>
</html>`

	t, _ := template.New("index").Parse(tmplStr)
	t.Execute(w, currentStatus)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 303)
		return
	}

	file, _, err := r.FormFile("torrent")
	if err != nil {
		currentStatus.Message = "Error: " + err.Error()
		http.Redirect(w, r, "/", 303)
		return
	}
	defer file.Close()

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

	params := engine.ClientParams{OutputPath: spec.Info.Name}
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
			currentStatus.Message = fmt.Sprintf("Downloading... %.1f%%", currentStatus.Percent)
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
