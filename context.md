## FlashDrop - Quick Local Network File Transfer

**What it is:** A simple Go-based HTTP server for transferring files from Mac to Android phone over local network.

**The Problem:** Need a quick way to transfer files/folders from MacBook to OnePlus Nord without cables, cloud services, or complex setup.

**The Solution:** 
- Run `flashdrop <directory_path>` on Mac
- Server auto-detects Mac's IP (`ipconfig getifaddr en0`)
- Creates beautiful mobile-friendly web interface at `http://[mac-ip]:8000`
- Phone visits URL, sees download button
- Click downloads → server zips directory → sends to phone
- Automatic cleanup of temp files

**Key Features:**
- Single executable, no dependencies
- Auto IP detection and display
- Responsive web UI with gradients/animations
- Smart zip naming with timestamps
- Progress feedback and file size logging
- Error handling for missing directories
- Temporary file cleanup

**Usage:**
```bash
go build -o flashdrop flashdrop.go
./flashdrop ~/Documents
# Open http://192.168.1.x:8000 on phone
```

**Tech Stack:** Pure Go with standard library (html/template, archive/zip, net/http), no external dependencies.

Perfect for ad-hoc file transfers within home/office network! 🚀📱