# 🚀 AutoServer


---

## AutoServer

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

[Releases](https://github.com/its-tinybinary-idk/autoserver/releases) ページから自分のプラットフォーム用のバイナリを取ってくれ。

| プラットフォーム | バイナリ |
|----------|--------|
| 🐧 Linux (x86_64) | `autoserver-linux-amd64` |
| 🐧 Linux (ARM64) | `autoserver-linux-arm64` |
| 🪟 Windows (x86_64) | `autoserver-windows-amd64.exe` |
| 🍎 macOS (Intel) | `autoserver-macos-intel` |
| 🍎 macOS (Apple Silicon) | `autoserver-macos-arm64` |

---

## 🚀 クイックスタート

1. [Releases](https://github.com/its-tinybinary-idk/autoserver/releases) から自分の OS 用のバイナリをダウンロード
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

 使ってるもの

· Go — ランタイムそのもの
· skip2/go-qrcode 
— 純 Go の QR コード生成
· 2011 年の古 PC — 死を拒否したやつ

---

AutoServer で頭痛が減ったなら、リポジトリに ⭐ をポチってくれ。マジで助かる。

---