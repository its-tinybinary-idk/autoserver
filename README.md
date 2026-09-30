# 🚀 AutoServer

[🇯🇵 日本語](#-日本語) | [🇬🇧 English](#-english)

---

## 🇯🇵 日本語

純粋な Go で書かれた、ちっちゃくて依存ゼロのローカル Web サーバー。HTML/CSS/JS が入ったフォルダにバイナリをポイっと置いて実行するだけで、同じ Wi-Fi 上のどの端末とでもすぐ共有できる。

設定ファイルなし。セットアップなし。`python -m http.server` で祈る必要なし。

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)
![License](https://img.shields.io/badge/license-MIT-green)

---

## ✨ 機能

- 📂 **コンテンツ自動検出** — HTML 1個でも、CSS/JS/画像入りのフォルダでも、勝手に判断する
- 🎛 **ダッシュボード UI** — ポート `9090` で自動で開くキレイなコントロールパネル
- 📱 **QR コード共有** — スマホでスキャン、IP アドレスを打つ必要なし
- 📋 **クリップボード自動コピー** — 頼む前にネットワーク URL がクリップボードに入ってる
- 🔒 **依存ゼロ** — CGO なし、ランタイムなし、Node なし、Python なし。バイナリ 1 個、以上。
- 🌍 **クロスプラットフォーム** — Windows、macOS、Linux、ARM。同じコードベースで、各 OS ネイティブのバイナリ。
- ⚡ **約 5 MB のバイナリ** — だいたいの npm モジュールより小さい
- 🛑 **Graceful shutdown** — Ctrl+C か停止ボタンで、ゾンビプロセスなし
- 🔐 **パストラバーサル対策** — 共有ネットワークや公共ネットワークでも安心

---

## 📦 ダウンロード

[Releases](../../releases) ページから自分のプラットフォーム用のバイナリを取ってくれ。

| プラットフォーム | バイナリ |
|----------|--------|
| 🐧 Linux (x86_64) | `autoserver-linux-amd64` |
| 🐧 Linux (ARM64) | `autoserver-linux-arm64` |
| 🪟 Windows (x86_64) | `autoserver-windows-amd64.exe` |
| 🍎 macOS (Intel) | `autoserver-macos-intel` |
| 🍎 macOS (Apple Silicon) | `autoserver-macos-arm64` |

---

## 🚀 クイックスタート

1. [Releases](../../releases) から自分の OS 用のバイナリをダウンロード
2. Linux/macOS なら実行権限を付ける:

   ```sh
   chmod +x autoserver-*
   ```

3. HTML ファイルが入ったフォルダにポイっと置く
4. 実行する（ダブルクリック、またはターミナルから ./autoserver-*）
5. ブラウザでダッシュボードが開く → スマホで QR コードをスキャン 📱

以上。設定なし、フラグなし、.env ファイルなし。

---

🎯 仕組み

AutoServer が起動すると：

1. os.Executable() で自分の場所を探す
2. そのフォルダ内の .html/.htm ファイルとサブフォルダをスキャン
3. 適切な配信モードを選ぶ：
   · 単一ファイルモード → HTML 1 個、サブフォルダ 0 → そのファイルを / で配信
   · フォルダモード → それ以外 → フォルダ全体をアセット込みで配信
4. HTTP サーバーを 2 つ並行で起動：
   · :8080 → 実際のコンテンツ
   · :9090 → コントロールダッシュボード
5. ネットワーク URL をクリップボードにコピー
6. デフォルトブラウザでダッシュボードを開く

ダッシュボードでできること：

· ステータスとモードを確認
· ローカル/ネットワーク URL をワンクリックでコピー
· スマホで QR コードをスキャン
· ターミナルを触らずにサーバーを停止

---

🔨 ソースからビルド

Go 1.24 以上が必要。

```sh
git clone https://github.com/its-tinybinary-idk/autoserver.git
cd autoserver
go build -o autoserver .
```

実行：

```sh
./autoserver
```

npm install なし。pip install なし。40 分の Rust コンパイルなし。Go の性格は「俺の時間を無駄にするな」、そこはリスペクトしてる。

---

🌍 クロスコンパイル

どのプラットフォームからでも、どのプラットフォーム向けでもビルド可能。ツールチェーン不要：

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o autoserver.exe
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o autoserver-mac
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o autoserver-linux
```

3 コマンド。5 プラットフォーム。苦しみなし。

---

🧰 動作環境

· ネットワーク：共有したい端末と同じ Wi-Fi

---

🐛 既知の制限

· QR コードライブラリは初回アクセス時に CDN から読み込むので、ダッシュボードは初回だけネットが必要。オフライン QR 生成はロードマップ上にある。

· ファイアウォールが 8080 と 9090 をブロックすることがある：

  · Linux: sudo ufw allow 8080 && sudo ufw allow 9090

  · Windows: プロンプトが出たら Windows Defender ファイアウォールでアプリを許可

---

🙏 使ってるもの

· Go — ランタイムそのもの
· skip2/go-qrcode 
— 純 Go の QR コード生成
· 2011 年の古 PC — 死を拒否したやつ

---

AutoServer で頭痛が減ったなら、リポジトリに ⭐ をポチってくれ。マジで助かる。

---

## 🇬🇧 English

A tiny, zero-dependency local web server written in pure Go. Drop the binary into any folder with HTML/CSS/JS, run it, and instantly share it with any device on your Wi-Fi.

No config files. No setup. No `python -m http.server` and hoping for the best.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)
![License](https://img.shields.io/badge/license-MIT-green)

---

## ✨ Features

- 📂 **Auto-detects content** — single HTML file or full folder with CSS/JS/images, it figures it out
- 🎛 **Dashboard UI** — clean control panel that opens automatically on port `9090`
- 📱 **QR code sharing** — scan with your phone, skip typing IP addresses
- 📋 **Clipboard auto-copy** — the network URL is on your clipboard before you even ask
- 🔒 **Zero dependencies** — no CGO, no runtime, no Node, no Python. One binary, that's it.
- 🌍 **Cross-platform** — Windows, macOS, Linux, ARM. Same codebase, native binaries for each.
- ⚡ **~5 MB binary** — smaller than most of your npm modules
- 🛑 **Graceful shutdown** — Ctrl+C or the stop button, no zombie processes
- 🔐 **Path traversal protection** — safe even on shared or public networks

---

## 📦 Downloads

Grab the binary for your platform from the [Releases](../../releases) page.

| Platform | Binary |
|----------|--------|
| 🐧 Linux (x86_64) | `autoserver-linux-amd64` |
| 🐧 Linux (ARM64) | `autoserver-linux-arm64` |
| 🪟 Windows (x86_64) | `autoserver-windows-amd64.exe` |
| 🍎 macOS (Intel) | `autoserver-macos-intel` |
| 🍎 macOS (Apple Silicon) | `autoserver-macos-arm64` |

---

## 🚀 Quick Start

1. Download the binary for your OS from [Releases](../../releases)
2. On Linux/macOS, make it executable:

   ```sh
   chmod +x autoserver-*
   ```

3. Drop it into a folder containing your HTML files
4. Run it (double-click, or ./autoserver-* from the terminal)
5. The dashboard opens in your browser → scan the QR code with your phone 📱

That's it. No configuration, no flags, no .env file.

---

🎯 How It Works

When AutoServer starts, it:

1. Finds its own location using os.Executable()
2. Scans that folder for .html/.htm files and subfolders
3. Picks the right serving mode:
   · Single-file mode → 1 HTML file, 0 subfolders → serves that file at /
   · Folder mode → anything else → serves the entire folder, assets included
4. Starts two HTTP servers side by side:
   · :8080 → your actual content
   · :9090 → the control dashboard
5. Copies the network URL to your clipboard
6. Opens the dashboard in your default browser

The dashboard lets you:

· See the status and mode
· Copy the local or network URL with one click
· Scan a QR code with your phone
· Stop the server without touching the terminal

---

🔨 Build From Source

Requires Go 1.24 or newer.

```sh
git clone https://github.com/its-tinybinary-idk/autoserver.git
cd autoserver
go build -o autoserver .
```

Run:

```sh
./autoserver
```

No npm install. No pip install. No 40-minute Rust compile. Go's whole personality is "stop wasting my time," and we respect it for that.

---

🌍 Cross-compile

Build for any platform from any platform, no toolchains required:

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o autoserver.exe
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o autoserver-mac
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o autoserver-linux
```

Three commands. Five platforms. No suffering.

---

🧰 System Requirements

· Network: Same Wi-Fi as the device you're sharing with

---

🐛 Known Limitations

· The QR code library loads from a CDN on first visit, so the dashboard needs internet the first time. Offline QR generation is on the roadmap.

· Firewall may block ports 8080 and 9090 on some systems:

  · Linux: sudo ufw allow 8080 && sudo ufw allow 9090

  · Windows: Allow the app through Windows Defender Firewall when prompted

---

🙏 Built With

· Go — the entire runtime
· skip2/go-qrcode 
— pure-Go QR code generation
· An old 2011 PC that refused to die

---

If AutoServer saved you a headache, drop a ⭐ on the repo. It genuinely helps.

```  
