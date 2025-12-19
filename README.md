# GoRunner


**GoRunner** is a high-performance, cross-platform desktop utility designed to automate Go testing workflows. It monitors your source code in real-time and triggers relevant test suites instantly upon file changes, facilitating a seamless TDD (Test-Driven Development) cycle.

<img width="1005" height="764" alt="image" src="https://github.com/user-attachments/assets/37a30ec1-86be-41f1-9e1a-c43305f8df1c" />

---

## 🚀 Key Features

- **Real-time Filesystem Observation**: Powered by `fsnotify` for instantaneous reaction to file modifications.
- **Intelligent Filtering**: Built-in support for exclusion lists (e.g., `.git`, `node_modules`, `vendor`) and customizable file glob masks.
- **Native OS Integration**: Leveraging PowerShell-based toast notifications on Windows for test results.
- **Automated Error Handling**: Optional "Auto-Copy Errors" feature that parses test failures and populates the system clipboard with the stack trace.
- **Multilingual Support**: Fully localized UI available in English and Russian.
- **Always-on-Top Mode**: Keeps the test console visible during heavy refactoring sessions.
- **Modern UI/UX**: Reactive frontend built with Vue 3 and Element Plus.

---

## 🛠 Technical Stack

- **Backend**: Go 1.24.5
- **Frontend**: Vue.js 3 (Composition API), Vite, Element Plus.
- **Bridge**: Wails v2 (High-speed IPC, Go-to-JS bindings).
- **Watcher**: Low-level events via `github.com/fsnotify/fsnotify`.

---

## ⚡ Quick Start

### Prerequisites
- **Go**: 1.24.5 or higher.
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`.
- **Node.js**: v18+ and npm (required for frontend assets).

### Building from Source
For Windows environments, use the provided build script:
```powershell
./build.bat
```
Alternatively, execute the Wails build command:
```bash
wails build -ldflags="-s -w"
```

### 📥 Installation
The easiest way to get started is to download the latest compiled executable from the **[GitHub Releases](https://github.com/smalloff/goruner/releases)** page.
