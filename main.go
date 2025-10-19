package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type MediaItem struct {
	ID          int       `db:"id" json:"id"`
	Path        string    `db:"path" json:"path"`
	Filename    string    `db:"filename" json:"filename"`
	Size        int64     `db:"size" json:"size"`
	Type        string    `db:"type" json:"type"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type App struct {
	DB *sqlx.DB
}

var supportedExtensions = map[string]string{
	".mp4":  "video",
	".avi":  "video",
	".mkv":  "video",
	".mov":  "video",
	".wmv":  "video",
	".flv":  "video",
	".webm": "video",
	".jpg":  "image",
	".jpeg": "image",
	".png":  "image",
	".gif":  "image",
	".webp": "image",
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)

	log.Info("Starting Media Organizer MVP...")

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	app := &App{DB: db}

	// Setup router
	r := chi.NewRouter()

	// API routes
	r.Get("/api/media", app.getMediaItems)
	r.Post("/api/scan", app.scanDirectory)
	r.Get("/api/stats", app.getStats)

	// Serve static files
	r.Get("/", serveIndex)
	r.Get("/static/*", http.NotFound)

	log.Info("Server starting on http://localhost:9999")
	log.Info("Open your browser and navigate to http://localhost:9999")
	http.ListenAndServe(":9999", r)
}

func initDB() (*sqlx.DB, error) {
	// Create data directory if it doesn't exist
	os.MkdirAll("./data", 0755)

	db, err := sqlx.Connect("sqlite3", "./data/media.db")
	if err != nil {
		return nil, err
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS media (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE,
		filename TEXT NOT NULL,
		size INTEGER NOT NULL,
		type TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_type ON media(type);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	log.Info("Database initialized successfully")
	return db, nil
}

func (app *App) getMediaItems(w http.ResponseWriter, r *http.Request) {
	mediaType := r.URL.Query().Get("type")

	var items []MediaItem
	var err error

	if mediaType != "" {
		err = app.DB.Select(&items, "SELECT * FROM media WHERE type = ? ORDER BY created_at DESC", mediaType)
	} else {
		err = app.DB.Select(&items, "SELECT * FROM media ORDER BY created_at DESC")
	}

	if err != nil {
		log.Error("Failed to fetch media items:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (app *App) scanDirectory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	log.Infof("Starting scan of directory: %s", req.Path)

	count := 0
	err := filepath.Walk(req.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		mediaType, ok := supportedExtensions[ext]
		if !ok {
			return nil
		}

		// Check if file already exists
		var existing int
		err = app.DB.Get(&existing, "SELECT COUNT(*) FROM media WHERE path = ?", path)
		if err == nil && existing > 0 {
			return nil
		}

		media := MediaItem{
			Path:     path,
			Filename: info.Name(),
			Size:     info.Size(),
			Type:     mediaType,
		}

		_, err = app.DB.NamedExec(
			"INSERT INTO media (path, filename, size, type) VALUES (:path, :filename, :size, :type)",
			media,
		)
		if err != nil {
			log.Warnf("Failed to insert media item %s: %v", path, err)
		} else {
			count++
		}

		return nil
	})

	if err != nil {
		log.Error("Failed to scan directory:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("Scan complete. Added %d new items", count)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   count,
		"message": fmt.Sprintf("Successfully scanned and added %d items", count),
	})
}

func (app *App) getStats(w http.ResponseWriter, r *http.Request) {
	var stats struct {
		Total  int `db:"total"`
		Videos int `db:"videos"`
		Images int `db:"images"`
	}

	err := app.DB.Get(&stats.Total, "SELECT COUNT(*) FROM media")
	if err != nil && err != sql.ErrNoRows {
		log.Error("Failed to get total count:", err)
	}

	err = app.DB.Get(&stats.Videos, "SELECT COUNT(*) FROM media WHERE type = 'video'")
	if err != nil && err != sql.ErrNoRows {
		log.Error("Failed to get video count:", err)
	}

	err = app.DB.Get(&stats.Images, "SELECT COUNT(*) FROM media WHERE type = 'image'")
	if err != nil && err != sql.ErrNoRows {
		log.Error("Failed to get image count:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html, _ := ioutil.ReadFile("index.html")
	if html == nil {
		w.Write([]byte(indexHTML))
	} else {
		w.Write(html)
	}
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Media Organizer MVP</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
        }

        header {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }

        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 32px;
        }

        .subtitle {
            color: #666;
            font-size: 14px;
        }

        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            text-align: center;
        }

        .stat-number {
            font-size: 36px;
            font-weight: bold;
            color: #667eea;
            margin-bottom: 5px;
        }

        .stat-label {
            color: #666;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }

        .controls {
            background: white;
            padding: 25px;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }

        .scan-section {
            display: flex;
            gap: 10px;
            align-items: center;
            flex-wrap: wrap;
        }

        input[type="text"] {
            flex: 1;
            min-width: 200px;
            padding: 12px 15px;
            border: 2px solid #e0e0e0;
            border-radius: 5px;
            font-size: 14px;
            transition: border-color 0.3s;
        }

        input[type="text"]:focus {
            outline: none;
            border-color: #667eea;
        }

        button {
            padding: 12px 24px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 5px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.3s;
        }

        button:hover {
            background: #5568d3;
        }

        button:disabled {
            background: #ccc;
            cursor: not-allowed;
        }

        .filter-buttons {
            display: flex;
            gap: 10px;
            margin-top: 15px;
        }

        .filter-btn {
            padding: 8px 16px;
            background: #f0f0f0;
            color: #333;
            border: 2px solid transparent;
            border-radius: 5px;
            font-size: 13px;
            cursor: pointer;
            transition: all 0.3s;
        }

        .filter-btn.active {
            background: #667eea;
            color: white;
            border-color: #667eea;
        }

        .media-grid {
            background: white;
            padding: 25px;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }

        .media-list {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }

        .media-item {
            background: #f9f9f9;
            padding: 15px;
            border-radius: 8px;
            border: 2px solid #e0e0e0;
            transition: all 0.3s;
        }

        .media-item:hover {
            border-color: #667eea;
            box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
        }

        .media-type {
            display: inline-block;
            padding: 4px 10px;
            background: #667eea;
            color: white;
            border-radius: 12px;
            font-size: 11px;
            text-transform: uppercase;
            font-weight: 600;
            margin-bottom: 10px;
        }

        .media-type.image {
            background: #48bb78;
        }

        .media-filename {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
            word-break: break-word;
        }

        .media-path {
            font-size: 12px;
            color: #666;
            margin-bottom: 8px;
            word-break: break-all;
        }

        .media-size {
            font-size: 12px;
            color: #999;
        }

        .message {
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
            display: none;
        }

        .message.success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }

        .message.error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }

        .message.show {
            display: block;
        }

        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #999;
        }

        .empty-state svg {
            width: 80px;
            height: 80px;
            margin-bottom: 20px;
            opacity: 0.3;
        }

        .loading {
            text-align: center;
            padding: 40px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>📁 Media Organizer MVP</h1>
            <p class="subtitle">Simplified version of Stash - Organize and browse your media collection</p>
        </header>

        <div class="stats">
            <div class="stat-card">
                <div class="stat-number" id="totalCount">0</div>
                <div class="stat-label">Total Items</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="videoCount">0</div>
                <div class="stat-label">Videos</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="imageCount">0</div>
                <div class="stat-label">Images</div>
            </div>
        </div>

        <div class="controls">
            <h3 style="margin-bottom: 15px; color: #333;">Scan Directory</h3>
            <div id="message" class="message"></div>
            <div class="scan-section">
                <input type="text" id="scanPath" placeholder="Enter directory path (e.g., /path/to/media)" />
                <button id="scanBtn" onclick="scanDirectory()">🔍 Scan</button>
            </div>
            <div class="filter-buttons">
                <button class="filter-btn active" onclick="filterMedia('')">All</button>
                <button class="filter-btn" onclick="filterMedia('video')">Videos</button>
                <button class="filter-btn" onclick="filterMedia('image')">Images</button>
            </div>
        </div>

        <div class="media-grid">
            <h3 style="color: #333; margin-bottom: 10px;">Media Library</h3>
            <div id="mediaList" class="media-list">
                <div class="loading">Loading...</div>
            </div>
        </div>
    </div>

    <script>
        let currentFilter = '';

        async function loadStats() {
            try {
                const response = await fetch('/api/stats');
                const stats = await response.json();
                document.getElementById('totalCount').textContent = stats.total || 0;
                document.getElementById('videoCount').textContent = stats.videos || 0;
                document.getElementById('imageCount').textContent = stats.images || 0;
            } catch (error) {
                console.error('Failed to load stats:', error);
            }
        }

        async function loadMedia(type = '') {
            try {
                const url = type ? `/api/media?type=${type}` : '/api/media';
                const response = await fetch(url);
                const media = await response.json();
                displayMedia(media);
            } catch (error) {
                console.error('Failed to load media:', error);
                document.getElementById('mediaList').innerHTML = '<div class="empty-state">Failed to load media</div>';
            }
        }

        function displayMedia(media) {
            const mediaList = document.getElementById('mediaList');
            
            if (!media || media.length === 0) {
                mediaList.innerHTML = `
                    <div class="empty-state">
                        <svg fill="currentColor" viewBox="0 0 20 20">
                            <path fill-rule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clip-rule="evenodd"></path>
                        </svg>
                        <h3>No media items found</h3>
                        <p>Scan a directory to add media to your library</p>
                    </div>
                `;
                return;
            }

            mediaList.innerHTML = media.map(item => `
                <div class="media-item">
                    <span class="media-type ${item.type}">${item.type}</span>
                    <div class="media-filename">${item.filename}</div>
                    <div class="media-path">${item.path}</div>
                    <div class="media-size">${formatSize(item.size)}</div>
                </div>
            `).join('');
        }

        function formatSize(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
        }

        async function scanDirectory() {
            const path = document.getElementById('scanPath').value;
            if (!path) {
                showMessage('Please enter a directory path', 'error');
                return;
            }

            const btn = document.getElementById('scanBtn');
            btn.disabled = true;
            btn.textContent = '⏳ Scanning...';

            try {
                const response = await fetch('/api/scan', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ path })
                });

                const result = await response.json();
                
                if (result.success) {
                    showMessage(result.message, 'success');
                    await loadStats();
                    await loadMedia(currentFilter);
                } else {
                    showMessage('Scan failed', 'error');
                }
            } catch (error) {
                showMessage('Failed to scan directory: ' + error.message, 'error');
            } finally {
                btn.disabled = false;
                btn.textContent = '🔍 Scan';
            }
        }

        function filterMedia(type) {
            currentFilter = type;
            document.querySelectorAll('.filter-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            event.target.classList.add('active');
            loadMedia(type);
        }

        function showMessage(text, type) {
            const messageDiv = document.getElementById('message');
            messageDiv.textContent = text;
            messageDiv.className = `message ${type} show`;
            setTimeout(() => {
                messageDiv.classList.remove('show');
            }, 5000);
        }

        // Load initial data
        loadStats();
        loadMedia();

        // Refresh stats every 30 seconds
        setInterval(loadStats, 30000);
    </script>
</body>
</html>`
