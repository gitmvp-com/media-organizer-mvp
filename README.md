# Media Organizer MVP

A simplified MVP version of [Stash](https://github.com/stashapp/stash) - a self-hosted media organizer with a web interface.

## Overview

This MVP focuses on the core functionality of Stash: organizing and browsing media files through a simple web interface. It provides:

- 📁 **Directory Scanning**: Scan directories to automatically index media files
- 🎬 **Media Management**: Support for video and image formats
- 📊 **Statistics**: View counts of your media library
- 🔍 **Filtering**: Filter by media type (videos/images)
- 🌐 **Web Interface**: Clean, modern UI accessible from any browser

## Features

### Core Features (MVP)

- ✅ SQLite database for media metadata storage
- ✅ HTTP server with REST API
- ✅ Directory scanning for videos and images
- ✅ Media library browser
- ✅ Basic statistics dashboard
- ✅ Filter by media type

### Not Included (Simplified for MVP)

- ❌ Authentication/Authorization
- ❌ GraphQL API
- ❌ Metadata scraping
- ❌ Tagging system
- ❌ Performer/Studio management
- ❌ Video streaming/transcoding
- ❌ Image thumbnails
- ❌ Plugins/Extensions

## Tech Stack

### Backend
- **Go 1.19** - Main programming language
- **chi** v4.0.2 - HTTP router
- **sqlx** v1.3.1 - SQL extensions
- **sqlite3** v1.14.7 - Database
- **logrus** v1.8.1 - Logging

### Frontend
- **Vanilla JavaScript** - No framework dependencies
- **HTML5 & CSS3** - Modern responsive design

## Installation

### Prerequisites

- Go 1.19 or higher
- GCC compiler (for SQLite)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/gitmvp-com/media-organizer-mvp.git
cd media-organizer-mvp

# Download dependencies
go mod download

# Build the application
go build -o media-organizer

# Run the application
./media-organizer
```

## Usage

### Starting the Server

```bash
./media-organizer
```

The server will start on `http://localhost:9999` by default.

### Scanning Directories

1. Open your browser and navigate to `http://localhost:9999`
2. Enter a directory path in the scan input field
3. Click the "🔍 Scan" button
4. The application will recursively scan the directory and add supported media files

### Supported Formats

**Videos:**
- .mp4, .avi, .mkv, .mov, .wmv, .flv, .webm

**Images:**
- .jpg, .jpeg, .png, .gif, .webp

### API Endpoints

#### Get Media Items
```
GET /api/media
GET /api/media?type=video
GET /api/media?type=image
```

#### Scan Directory
```
POST /api/scan
Content-Type: application/json

{
  "path": "/path/to/media"
}
```

#### Get Statistics
```
GET /api/stats
```

Response:
```json
{
  "total": 150,
  "videos": 100,
  "images": 50
}
```

## Project Structure

```
.
├── main.go           # Main application code
├── go.mod            # Go module definition
├── go.sum            # Go module checksums
├── README.md         # This file
└── data/             # Created at runtime
    └── media.db      # SQLite database
```

## Configuration

The application uses minimal configuration:

- **Port**: `9999` (hardcoded in main.go)
- **Database**: `./data/media.db` (SQLite)
- **Log Level**: `Info`

## Development

### Running in Development Mode

```bash
go run main.go
```

### Building for Production

```bash
# Linux/macOS
go build -ldflags="-s -w" -o media-organizer

# Windows
go build -ldflags="-s -w" -o media-organizer.exe

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o media-organizer-linux
GOOS=darwin GOARCH=amd64 go build -o media-organizer-macos
GOOS=windows GOARCH=amd64 go build -o media-organizer.exe
```

## Differences from Original Stash

This MVP is a **significantly simplified** version of Stash:

| Feature | Stash | This MVP |
|---------|-------|----------|
| Backend | Go + GraphQL | Go + REST API |
| Frontend | React | Vanilla JS |
| Database | SQLite with migrations | Simple SQLite |
| Auth | Yes | No |
| Metadata | Scrapers + StashDB | None |
| Media Types | Videos, Images, Galleries | Videos, Images |
| Tagging | Advanced tagging system | None |
| Performers | Full management | None |
| Streaming | FFmpeg transcoding | None |
| Thumbnails | Auto-generated | None |
| Plugins | Plugin system | None |

## Limitations

- No authentication - anyone with network access can use it
- No metadata scraping or external integrations
- No video playback or image viewing in the interface
- No thumbnail generation
- Limited to local file system scanning
- No concurrent scan protection
- Basic error handling

## Future Enhancements

Potential features to add:

- [ ] In-browser media viewer
- [ ] Thumbnail generation
- [ ] Basic tagging system
- [ ] Search functionality
- [ ] Configuration file support
- [ ] Docker support
- [ ] Database migrations
- [ ] Better error handling
- [ ] Progress tracking for scans
- [ ] Duplicate detection

## Contributing

This is an MVP project for learning purposes. Feel free to fork and extend it!

## License

This project is inspired by [Stash](https://github.com/stashapp/stash) but is an independent implementation.
Refer to the original Stash project for their licensing terms.

## Acknowledgments

- **Stash Team**: For creating the original amazing media organizer
- **Go Community**: For the excellent libraries used in this project

## Resources

- [Original Stash Repository](https://github.com/stashapp/stash)
- [Stash Website](https://stashapp.cc)
- [Stash Documentation](https://docs.stashapp.cc)

---

**Note**: This is a learning project and MVP. For production use, please use the original [Stash](https://github.com/stashapp/stash) which is feature-complete and actively maintained.
