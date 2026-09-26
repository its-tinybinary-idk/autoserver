package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/skip2/go-qrcode"
)

// ============ GLOBALS ============

var (
	hostedPath    string
	hostedMode    string // "single" or "folder"
	hostedFiles   []string
	contentServer *http.Server
	dashServer    *http.Server
	startTime     time.Time
	mu            sync.Mutex
)

// ============ NETWORK HELPERS ============

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// ============ BROWSER LAUNCHER ============

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// ============ CLIPBOARD ============

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "echo "+text+"| clip")
	case "darwin":
		cmd = exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
	default:
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
			cmd.Stdin = strings.NewReader(text)
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
			cmd.Stdin = strings.NewReader(text)
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
			cmd.Stdin = strings.NewReader(text)
		} else {
			return
		}
	}
	_ = cmd.Run()
}

// ============ FOLDER SCANNER ============

func scanFolder(rootDir string) (singleMode bool, singleFile string, htmlCount, subfolderCount int) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return false, "", 0, 0
	}

	var htmlFiles []string
	for _, e := range entries {
		if e.IsDir() {
			subfolderCount++
		} else {
			lower := strings.ToLower(e.Name())
			if strings.HasSuffix(lower, ".html") || strings.HasSuffix(lower, ".htm") {
				htmlFiles = append(htmlFiles, e.Name())
			}
		}
	}

	htmlCount = len(htmlFiles)
	hostedFiles = htmlFiles

	if htmlCount == 1 && subfolderCount == 0 {
		return true, htmlFiles[0], 1, 0
	}
	return false, "", htmlCount, subfolderCount
}

// ============ SERVER STARTERS ============

func startContentServer(rootDir string, singleMode bool, singleFile string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Security: block directory traversal attempts
		cleanPath := filepath.Clean(r.URL.Path)
		if strings.Contains(cleanPath, "..") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if singleMode && (r.URL.Path == "/" || r.URL.Path == "") {
			http.ServeFile(w, r, filepath.Join(rootDir, singleFile))
			return
		}
		http.FileServer(http.Dir(rootDir)).ServeHTTP(w, r)
	})

	// Tiny health endpoint
	mux.HandleFunc("/__health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	contentServer = &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		if err := contentServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Content server error:", err)
		}
	}()
}

func startDashboardServer() {
	mux := http.NewServeMux()

	// Dashboard page
	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, renderDashboard())
	})

	// QR code PNG
	mux.HandleFunc("/qr", func(w http.ResponseWriter, r *http.Request) {
		ip := getLocalIP()
		networkURL := "http://" + ip + ":8080"
		png, err := qrcode.Encode(networkURL, qrcode.Medium, 256)
		if err != nil {
			http.Error(w, "QR error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(png)
	})

	// Stats API (JSON)
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		uptime := time.Since(startTime).Round(time.Second)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"uptime":"%s","mode":"%s","folder":"%s","ip":"%s"}`,
			uptime.String(), hostedMode, escapeJSON(hostedPath), getLocalIP())
	})

	// Shutdown API
	mux.HandleFunc("/api/shutdown", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Shutting down...")
		go func() {
			time.Sleep(500 * time.Millisecond)
			gracefulShutdown()
		}()
	})

	// Root redirects to dashboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})

	dashServer = &http.Server{Addr: ":9090", Handler: mux}
	go func() {
		if err := dashServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Dashboard server error:", err)
		}
	}()
}

// ============ GRACEFUL SHUTDOWN ============

func gracefulShutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if contentServer != nil {
		contentServer.Shutdown(ctx)
	}
	if dashServer != nil {
		dashServer.Shutdown(ctx)
	}
	fmt.Println("\n🛑 Server stopped.")
	os.Exit(0)
}

// ============ UTILITIES ============

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func pause() {
	if runtime.GOOS == "windows" {
		fmt.Println("\nPress Enter to exit...")
		fmt.Scanln()
	}
}

// ============ MAIN ============

func main() {
	startTime = time.Now()

	// Find binary location
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error finding binary:", err)
		pause()
		os.Exit(1)
	}
	rootDir := filepath.Dir(exePath)
	if resolved, err := filepath.EvalSymlinks(rootDir); err == nil {
		rootDir = resolved
	}
	hostedPath = rootDir

	// Scan folder
	singleMode, singleFile, htmlCount, subfolderCount := scanFolder(rootDir)
	if singleMode {
		hostedMode = "Single HTML: " + singleFile
	} else {
		hostedMode = fmt.Sprintf("Folder (%d HTML, %d subfolders)", htmlCount, subfolderCount)
	}

	// Start servers
	startContentServer(rootDir, singleMode, singleFile)
	startDashboardServer()

	// Give servers a moment to bind
	time.Sleep(150 * time.Millisecond)

	// Compute URLs
	ip := getLocalIP()
	localURL := "http://localhost:8080"
	networkURL := "http://" + ip + ":8080"
	dashURL := "http://localhost:9090/dashboard"

	// Copy network URL to clipboard
	copyToClipboard(networkURL)

	// Print banner
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║           🚀  AUTO SERVER  🚀                        ║")
	fmt.Println("╠══════════════════════════════════════════════════════╣")
	fmt.Printf("║  📂 Folder: %-42s ║\n", truncate(rootDir, 42))
	fmt.Printf("║  📄 Mode:   %-42s ║\n", truncate(hostedMode, 42))
	fmt.Println("╠══════════════════════════════════════════════════════╣")
	fmt.Printf("║  🎛  Dashboard: %-38s ║\n", dashURL)
	fmt.Printf("║  💻 Local:      %-38s ║\n", localURL)
	fmt.Printf("║  📱 Network:    %-38s ║\n", networkURL)
	fmt.Println("╠══════════════════════════════════════════════════════╣")
	fmt.Println("║  📋 Network URL copied to clipboard                  ║")
	fmt.Println("║  🌐 Opening dashboard in your browser...             ║")
	fmt.Println("║  ⏹  Close this window OR press Ctrl+C to stop.       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	// Auto-open browser to dashboard
	go func() {
		time.Sleep(300 * time.Millisecond)
		openBrowser(dashURL)
	}()

	// Handle Ctrl+C gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		gracefulShutdown()
	}()

	// Block forever
	select {}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "..." + s[len(s)-(max-3):]
}

// ============ DASHBOARD HTML ============

func renderDashboard() string {
	ip := getLocalIP()
	localURL := "http://localhost:8080"
	networkURL := "http://" + ip + ":8080"

	mode := hostedMode
	folder := hostedPath

	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Auto Server Dashboard</title>
	<style>
		* { box-sizing: border-box; }
		body {
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif;
			background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
			color: #eaeaea;
			margin: 0;
			padding: 40px 20px;
			min-height: 100vh;
			-webkit-font-smoothing: antialiased;
		}
		.container { max-width: 640px; margin: 0 auto; }
		h1 {
			text-align: center;
			font-size: 32px;
			margin: 0 0 8px;
			background: linear-gradient(90deg, #4ecca3, #7f9cf5);
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
			background-clip: text;
		}
		.subtitle { text-align: center; color: #888; font-size: 14px; margin-bottom: 30px; }
		.card {
			background: rgba(22, 33, 62, 0.7);
			backdrop-filter: blur(10px);
			border: 1px solid rgba(255,255,255,0.05);
			border-radius: 14px;
			padding: 25px;
			margin-bottom: 20px;
			box-shadow: 0 8px 32px rgba(0,0,0,0.3);
		}
		.card h2 {
			margin: 0 0 15px;
			font-size: 13px;
			color: #888;
			text-transform: uppercase;
			letter-spacing: 1.5px;
			font-weight: 600;
		}
		.status-row { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; }
		.status {
			display: inline-block;
			padding: 4px 12px;
			background: #4ecca3;
			color: #1a1a2e;
			border-radius: 20px;
			font-size: 11px;
			font-weight: bold;
			letter-spacing: 1px;
		}
		.status::before { content: "● "; animation: pulse 1.5s infinite; }
		@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.4; } }
		.info { display: flex; justify-content: space-between; font-size: 14px; color: #aaa; padding: 8px 0; border-bottom: 1px solid rgba(255,255,255,0.05); }
		.info:last-child { border-bottom: none; }
		.info strong { color: #eaeaea; font-weight: 500; text-align: right; max-width: 60%; word-break: break-all; }
		.url {
			font-size: 20px;
			font-weight: 600;
			color: #4ecca3;
			word-break: break-all;
			margin: 12px 0;
			padding: 14px;
			background: rgba(78, 204, 163, 0.1);
			border-radius: 8px;
			border-left: 3px solid #4ecca3;
			user-select: all;
			font-family: "SF Mono", Menlo, monospace;
		}
		.url.local { color: #7f9cf5; background: rgba(127, 156, 245, 0.1); border-left-color: #7f9cf5; }
		.btn-row { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 12px; }
		.btn {
			flex: 1;
			min-width: 120px;
			background: #4ecca3;
			color: #1a1a2e;
			border: none;
			padding: 12px 20px;
			border-radius: 8px;
			font-size: 14px;
			cursor: pointer;
			font-weight: 600;
			transition: all 0.2s;
			text-decoration: none;
			text-align: center;
			display: inline-block;
		}
		.btn:hover { background: #45b890; transform: translateY(-1px); }
		.btn.secondary { background: transparent; color: #7f9cf5; border: 1px solid #7f9cf5; }
		.btn.secondary:hover { background: rgba(127, 156, 245, 0.1); }
		.btn.danger { background: transparent; color: #ff6b6b; border: 1px solid #ff6b6b; }
		.btn.danger:hover { background: rgba(255, 107, 107, 0.1); }
		#qrcode {
			display: flex;
			justify-content: center;
			padding: 20px;
			background: #fff;
			border-radius: 12px;
			margin-top: 15px;
		}
		#qrcode img { display: block; border-radius: 8px; }
		.hint { font-size: 13px; color: #888; margin-top: 15px; line-height: 1.6; }
		.uptime { text-align: center; color: #666; font-size: 12px; margin-top: 20px; font-family: monospace; }
		.toast {
			position: fixed;
			bottom: 30px;
			left: 50%;
			transform: translateX(-50%) translateY(100px);
			background: #4ecca3;
			color: #1a1a2e;
			padding: 12px 24px;
			border-radius: 8px;
			font-weight: 600;
			transition: transform 0.3s;
			box-shadow: 0 8px 24px rgba(78, 204, 163, 0.4);
			z-index: 9999;
		}
		.toast.show { transform: translateX(-50%) translateY(0); }
	</style>
</head>
<body>
	<div class="container">
		<h1>🎛 Auto Server</h1>
		<p class="subtitle">Your personal file hosting dashboard</p>

		<div class="card">
			<h2>Server Status</h2>
			<div class="status-row">
				<span class="status">RUNNING</span>
			</div>
			<div class="info">
				<span>Mode</span>
				<strong>` + escapeHTML(mode) + `</strong>
			</div>
			<div class="info">
				<span>Folder</span>
				<strong style="font-size:12px;">` + escapeHTML(folder) + `</strong>
			</div>
			<div class="info">
				<span>Uptime</span>
				<strong id="uptime">--</strong>
			</div>
		</div>

		<div class="card">
			<h2>📱 Share with phone / other devices</h2>
			<div class="url" id="neturl">` + networkURL + `</div>
			<div id="qrcode"></div>
			<div class="btn-row">
				<button class="btn" onclick="copyUrl('` + networkURL + `')">📋 Copy URL</button>
			</div>
			<p class="hint">📸 Scan the QR code with your phone camera, or tap "Copy URL". Make sure your phone is on the <strong>same Wi-Fi network</strong>.</p>
		</div>

		<div class="card">
			<h2>💻 Local access (this computer)</h2>
			<div class="url local">` + localURL + `</div>
			<div class="btn-row">
				<button class="btn" onclick="copyUrl('` + localURL + `')">📋 Copy</button>
				<a class="btn secondary" href="` + localURL + `" target="_blank">🚀 Open Site</a>
			</div>
		</div>

		<div class="card">
			<h2>⚙️ Controls</h2>
			<div class="btn-row">
				<button class="btn danger" onclick="shutdown()">⏹ Stop Server</button>
			</div>
			<p class="hint">Stopping will close the server and the terminal window.</p>
		</div>

		<p class="uptime" id="footer">Auto Server • Pure Go • Zero deps</p>
	</div>

	<div class="toast" id="toast">Copied!</div>

	<script>
		// Copy helper
		function copyUrl(url) {
			navigator.clipboard.writeText(url).then(() => showToast('Copied: ' + url));
		}

		function showToast(msg) {
			const t = document.getElementById('toast');
			t.textContent = msg;
			t.classList.add('show');
			setTimeout(() => t.classList.remove('show'), 2000);
		}

		// Shutdown
		function shutdown() {
			if (confirm('Stop the server and close everything?')) {
				fetch('/api/shutdown').then(() => {
					document.body.innerHTML = '<div style="text-align:center;padding:100px;font-size:24px;color:#4ecca3;">✅ Server stopped. You can close this tab.</div>';
				});
			}
		}

		// Uptime poller
		function updateStats() {
			fetch('/api/stats').then(r => r.json()).then(data => {
				document.getElementById('uptime').textContent = data.uptime;
			}).catch(() => {});
		}
		setInterval(updateStats, 1000);
		updateStats();

		// Keyboard shortcut: press 's' to shutdown
		document.addEventListener('keydown', (e) => {
			if (e.key === 's' && !e.target.matches('input,textarea')) shutdown();
		});
	</script>
</body>
</html>`
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
