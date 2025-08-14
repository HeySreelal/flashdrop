package main

import (
	"archive/zip"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var targetDir string
var serverIP string

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FlashDrop 🚀</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            padding: 2rem;
            border-radius: 20px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            text-align: center;
            max-width: 400px;
            width: 90%;
        }
        h1 {
            color: #333;
            margin-bottom: 1rem;
            font-size: 2rem;
        }
        .info {
            background: #f8f9fa;
            padding: 1rem;
            border-radius: 10px;
            margin: 1rem 0;
            color: #666;
        }
        .download-btn {
            background: linear-gradient(45deg, #667eea, #764ba2);
            color: white;
            border: none;
            padding: 1rem 2rem;
            font-size: 1.1rem;
            border-radius: 50px;
            cursor: pointer;
            transition: transform 0.2s;
            margin-top: 1rem;
            font-weight: bold;
        }
        .download-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(0,0,0,0.2);
        }
        .footer {
            margin-top: 1rem;
            color: #999;
            font-size: 0.9rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>FlashDrop 🚀</h1>
        <div class="info">
            <strong>Directory:</strong> {{.Directory}}<br>
            <strong>Server IP:</strong> {{.ServerIP}}:8000
        </div>
        <button class="download-btn" onclick="downloadFiles()">
            📦 Download Files
        </button>
        <div class="footer">
            Click to zip and download the directory
        </div>
    </div>
    
    <script>
        function downloadFiles() {
            const btn = document.querySelector('.download-btn');
            btn.innerHTML = '⏳ Preparing download...';
            btn.disabled = true;
            
            window.location.href = '/download';
            
            setTimeout(() => {
                btn.innerHTML = '📦 Download Files';
                btn.disabled = false;
            }, 2000);
        }
    </script>
</body>
</html>
`

func getLocalIP() string {
	cmd := exec.Command("ipconfig", "getifaddr", "en0")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Could not get IP address: %v", err)
		return "localhost"
	}
	return strings.TrimSpace(string(output))
}

func zipDirectory(src, dst string) error {
	zipFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create zip header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Use forward slashes for zip paths (cross-platform compatibility)
		header.Name = filepath.ToSlash(relPath)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		}

		return nil
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("home").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Directory string
		ServerIP  string
	}{
		Directory: targetDir,
		ServerIP:  serverIP,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Create temporary zip file
	timestamp := time.Now().Format("20060102_150405")
	dirName := filepath.Base(targetDir)
	zipFileName := fmt.Sprintf("%s_%s.zip", dirName, timestamp)
	tempZipPath := filepath.Join(os.TempDir(), zipFileName)

	log.Printf("Creating zip file: %s", tempZipPath)

	// Create the zip
	err := zipDirectory(targetDir, tempZipPath)
	if err != nil {
		log.Printf("Error creating zip: %v", err)
		http.Error(w, "Error creating zip file", http.StatusInternalServerError)
		return
	}

	// Clean up temp file after sending
	defer func() {
		os.Remove(tempZipPath)
		log.Printf("Cleaned up temp file: %s", tempZipPath)
	}()

	// Open the zip file
	zipFile, err := os.Open(tempZipPath)
	if err != nil {
		log.Printf("Error opening zip file: %v", err)
		http.Error(w, "Error opening zip file", http.StatusInternalServerError)
		return
	}
	defer zipFile.Close()

	// Get file info for size
	fileInfo, err := zipFile.Stat()
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFileName))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Send the file
	log.Printf("Sending zip file: %s (%.2f MB)", zipFileName, float64(fileInfo.Size())/(1024*1024))
	_, err = io.Copy(w, zipFile)
	if err != nil {
		log.Printf("Error sending file: %v", err)
		return
	}

	log.Printf("Successfully sent zip file to client")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("🚀 FlashDrop - Quick File Transfer Server")
		fmt.Println("Usage: flashdrop <directory_path>")
		fmt.Println("Example: flashdrop /Users/john/Documents")
		os.Exit(1)
	}

	targetDir = os.Args[1]

	// Check if directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("❌ Error: Directory '%s' does not exist\n", targetDir)
		os.Exit(1)
	}

	// Get absolute path
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Printf("❌ Error getting absolute path: %v\n", err)
		os.Exit(1)
	}
	targetDir = absPath

	// Get local IP
	serverIP = getLocalIP()

	// Set up routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/download", downloadHandler)

	fmt.Println("🚀 FlashDrop Server Started!")
	fmt.Printf("📁 Directory: %s\n", targetDir)
	fmt.Printf("🌐 Server URL: http://%s:8000\n", serverIP)
	fmt.Printf("📱 Open this URL on your phone to download files\n")
	fmt.Println("⏹️  Press Ctrl+C to stop the server")

	log.Fatal(http.ListenAndServe(":8000", nil))
}
